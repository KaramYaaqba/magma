/**
 * Copyright 2004-present Facebook. All Rights Reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * @flow
 * @format
 */
'use strict';

import logging from '@fbcnms/logging';
import {checkWasmer} from './wasmer.js';
import {executePython, pythonHealthCheck} from './python.js';
import {executeQuickJs, quickJsHealthCheck} from './quickjs.js';
const logger = logging.getLogger(module);

const ConductorClient = require('conductor-client').default;
// properties
const conductorApiUrl =
  process.env.CONDUCTOR_API_URL || 'http://conductor-server:8080/api';
const maxRunner = process.env.MAX_RUNNER || 1;

const conductorClient = new ConductorClient({
  baseURL: conductorApiUrl,
});

function registerWasmWorker(workerSuffix, callback) {
  conductorClient.registerWatcher(
    'GLOBAL___' + workerSuffix,
    callback,
    {pollingIntervals: 1000, autoAck: true, maxRunner: maxRunner},
    true,
  );
}

async function createTaskResult(
  outputIsJson,
  outputData,
  stderr,
  updater,
  reasonForIncompletion,
) {
  // split stderr to array, remove empty lines
  const logs = stderr.split(/\r?\n/).filter(String);
  // unless outputIsJson is disabled or task already failed,
  // try to interpret result as JSON
  if (outputIsJson !== 'false' || reasonForIncompletion != null) {
    try {
      outputData.result = JSON.parse(outputData.result);
    } catch (e) {
      logs.push('Converting stdout to json failed');
      if (outputIsJson === 'true') {
        // if user specifies that this must be JSON, fail the task
        if (reasonForIncompletion == null) {
          reasonForIncompletion = 'Invalid JSON';
        }
      }
    }
  }
  const updaterFun =
    reasonForIncompletion == null ? updater.complete : updater.fail;
  logger.info('createTaskResult updating task', {
    outputData,
    logs,
    reasonForIncompletion,
  });
  await updaterFun({
    outputData,
    logs,
    reasonForIncompletion,
  });
}

async function checkAndRegister(wasmSuffix, healthCheckFn, executeFn) {
  const healthCheckStart = new Date();
  if (!(await healthCheckFn())) {
    logger.warn(`${wasmSuffix} healthcheck failed`);
  } else {
    logger.info(
      `${wasmSuffix} healthcheck OK in ${new Date() - healthCheckStart} ms`,
    );
  }
  registerWasmWorker(wasmSuffix, async (data, updater) => {
    logger.info(wasmSuffix + ' got new task', {inputData: data.inputData});
    const inputData = data.inputData;
    const args = inputData.args;
    const outputIsJson = inputData.outputIsJson;
    const scriptExpression = inputData.scriptExpression;
    try {
      const {stdout, stderr} = await executeFn(
        scriptExpression,
        args,
        inputData,
      );
      await createTaskResult(outputIsJson, {result: stdout}, stderr, updater);
    } catch (e) {
      logger.error('Task has failed', {
        killed: e.killed,
        code: e.code,
        signal: e.signal,
        cmd: e.cmd,
      });
      logger.info('Task has failed', {error: e});
      let reasonForIncompletion = 'Unknown reason';
      if (e.killed) {
        reasonForIncompletion = 'Timeout';
      } else if (e.code != null && e.code != 0) {
        reasonForIncompletion = 'Exited with error ' + e.code;
      }
      await createTaskResult(
        outputIsJson,
        {result: e.stdout},
        e.stderr,
        updater,
        reasonForIncompletion,
      );
    }
  });
}

async function registerTaskDefs() {
  const taskDefs = [
    {
      name: 'GLOBAL___js',
      type: 'SIMPLE',
      retryCount: 3,
      retryLogic: 'FIXED',
      retryDelaySeconds: 10,
      timeoutSeconds: 300,
      timeoutPolicy: 'TIME_OUT_WF',
      responseTimeoutSeconds: 180,
      ownerEmail: 'example@example.com',
    },
    {
      name: 'GLOBAL___py',
      type: 'SIMPLE',
      retryCount: 3,
      retryLogic: 'FIXED',
      retryDelaySeconds: 10,
      timeoutSeconds: 300,
      timeoutPolicy: 'TIME_OUT_WF',
      responseTimeoutSeconds: 180,
      ownerEmail: 'example@example.com',
    },
  ];
  await conductorClient.registerTaskDefs(taskDefs);
}

async function init() {
  await checkWasmer();

  await registerTaskDefs();

  const workers = new Map([
    ['js', {healthCheckFn: quickJsHealthCheck, executeFn: executeQuickJs}],
    ['py', {healthCheckFn: pythonHealthCheck, executeFn: executePython}],
  ]);

  for (const [wasmSuffix, {healthCheckFn, executeFn}] of workers) {
    try {
      await checkAndRegister(wasmSuffix, healthCheckFn, executeFn);
    } catch (error) {
      logger.warn('Error in checkAndRegister of ' + wasmSuffix, {error});
    }
  }
}

async function initWithRetry() {
  // auto reconnect is not supported by conductor-client,
  // retry on error here
  try {
    await init();
    return;
  } catch (error) {
    console.error('Got error, reconnecting', {error});
    setTimeout(initWithRetry, 1000);
  }
}

initWithRetry();
