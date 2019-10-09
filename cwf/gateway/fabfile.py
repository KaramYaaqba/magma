"""
Copyright (c) 2016-present, Facebook, Inc.
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree. An additional grant
of patent rights can be found in the PATENTS file in the same directory.
"""

import sys
from distutils.util import strtobool

from fabric.api import cd, env, execute, lcd, local, run, settings, sudo

sys.path.append('../../orc8r')
from tools.fab.hosts import ansible_setup, vagrant_setup

CWAG_ROOT = "$MAGMA_ROOT/cwf/gateway"
CWAG_INTEG_ROOT = "$MAGMA_ROOT/cwf/gateway/integ_tests"
LTE_AGW_ROOT = "../../lte/gateway"

TRF_SERVER_IP = "192.168.129.42"
TRF_SERVER_SUBNET = "192.168.129.0"
CWAG_BR_NAME = "cwag_br0"
CWAG_TEST_BR_NAME = "cwag_test_br0"


def integ_test(gateway_host=None, test_host=None, trf_host=None,
               destroy_vm="False"):
    """
    Run the integration tests. This defaults to running on local vagrant
    machines, but can also be pointed to an arbitrary host (e.g. amazon) by
    passing "address:port" as arguments

    gateway_host: The ssh address string of the machine to run the gateway
        services on. Formatted as "host:port". If not specified, defaults to
        the `cwag` vagrant box.

    test_host: The ssh address string of the machine to run the tests on
        on. Formatted as "host:port". If not specified, defaults to the
        `cwag_test` vagrant box.

    trf_host: The ssh address string of the machine to run the tests on
        on. Formatted as "host:port". If not specified, defaults to the
        `magma_trfserver` vagrant box.
    """

    destroy_vm = bool(strtobool(destroy_vm))

    # Setup the gateway: use the provided gateway if given, else default to the
    # vagrant machine
    if not gateway_host:
        vagrant_setup("cwag", destroy_vm)
    else:
        ansible_setup(gateway_host, "cwag", "cwag_dev.yml")

    execute(_run_unit_tests)
    execute(_set_cwag_configs)
    cwag_host_to_mac = execute(_get_cwag_br_mac)
    cwag_br_mac = cwag_host_to_mac[env.host_string]
    execute(_start_gateway)

    # Setup the trfserver: use the provided trfserver if given, else default to the
    # vagrant machine
    with lcd(LTE_AGW_ROOT):
        if not trf_host:
            vagrant_setup("magma_trfserver", destroy_vm)
        else:
            ansible_setup(trf_host, "trfserver", "magma_trfserver.yml")

    execute(_start_trfserver)

    # Run the tests: use the provided test machine if given, else default to
    # the vagrant machine
    if not test_host:
        vagrant_setup("cwag_test", destroy_vm)
    else:
        ansible_setup(test_host, "cwag_test", "cwag_test.yml")

    execute(_set_cwag_test_configs)
    execute(_set_cwag_test_networking, cwag_br_mac)
    execute(_start_ue_simulator)
    execute(_run_integ_tests)


def _set_cwag_configs():
    """ Set the necessary config overrides """

    with cd(CWAG_INTEG_ROOT):
        sudo('mkdir -p /var/opt/magma')
        sudo('mkdir -p /var/opt/magma/configs')
        sudo('cp gateway.mconfig /var/opt/magma/configs')
        sudo('cp sessiond.yml /var/opt/magma/configs')


def _get_cwag_br_mac():
    mac = run("cat /sys/class/net/%s/address" % CWAG_BR_NAME)
    return mac


def _set_cwag_test_configs():
    """ Set the necessary test configs """

    sudo('mkdir -p /etc/magma')
    # Create empty uesim config
    sudo('touch /etc/magma/uesim.yml')


def _set_cwag_test_networking(mac):
    # Don't error if route already exists
    with settings(warn_only=True):
        sudo('ip route add %s/24 dev %s proto static scope link' %
             (TRF_SERVER_SUBNET, CWAG_TEST_BR_NAME))
    sudo('arp -s %s %s' % (TRF_SERVER_IP, mac))


def _start_gateway():
    """ Starts the gateway """
    with cd(CWAG_ROOT + '/docker'):
        sudo(' docker-compose'
            ' -f docker-compose.yml'
            ' -f docker-compose.override.yml'
            ' -f docker-compose.integ-test.yml'
            ' build --parallel')
        sudo(' docker-compose'
            ' -f docker-compose.yml'
            ' -f docker-compose.override.yml'
            ' -f docker-compose.integ-test.yml'
            ' up -d')


def _start_ue_simulator():
    """ Starts the UE Sim Service """
    with cd(CWAG_ROOT + '/services/uesim/uesim'):
        run('tmux new -d \'go run main.go\'')


def _start_trfserver():
    """ Starts the traffic gen server"""
    run('nohup iperf3 -s -B %s > /dev/null &' % TRF_SERVER_IP, pty=False)


def _run_unit_tests():
    """ Run the cwag unit tests """

    with cd(CWAG_ROOT):
        run('make test')


def _run_integ_tests():
    """ Run the integration tests """

    with cd(CWAG_INTEG_ROOT):
        run('make integ_test')
