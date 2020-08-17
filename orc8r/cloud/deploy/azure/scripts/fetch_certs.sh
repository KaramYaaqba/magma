mkdir orc8r/cloud/helm/orc8r/charts/secrets/secrets/certs
cd orc8r/cloud/helm/orc8r/charts/secrets/secrets/certs

sudo apt install jq

echo "******Acquiring Sonarlte Certificate******"

echo "Downloading pfx (no password) from Key vault..."
secretId=$(az keyvault certificate show --vault-name $2 --name sonarlte-com | jq -r ".sid")
az keyvault secret download -f sonar-nopass.pfx --vault-name $2 --id $secretId --encoding base64

echo "Converting pfx (no password) to pem format..."
openssl pkcs12 -in sonar-nopass.pfx -out sonar-nopass.pem -nodes -password pass:""

echo "Converting pem to pfx with the password..."
openssl pkcs12 -export -out sonar-lte.pfx -in sonar-nopass.pem -password pass:"$1"

echo "Converting pfx (with password) to pem..."
openssl pkcs12 -in sonar-lte.pfx -nokeys -out rootCA.pem -nodes -password pass:"$1"

echo "Extracting private key from pfx..."
openssl pkcs12 -in sonar-lte.pfx -nocerts -out controller-encrypted.key -passin pass:"$1" -passout pass:"$1"

echo "Decrypting private key..."
openssl rsa -in controller-encrypted.key -out controller.key -passin pass:"$1"

echo "Extracting public certificate from pfx..."
openssl pkcs12 -in sonar-lte.pfx -clcerts -nokeys -out controller.crt -password pass:"$1"

echo "*******************************************"

echo "******Acquiring Certifier Certificate******"

echo "Downloading pfx (no password) from Key vault..."
secretId=$(az keyvault certificate show --vault-name $2 --name Certifier | jq -r ".sid")
az keyvault secret download -f certifier-nopass.pfx --vault-name $2 --id $secretId --encoding base64

echo "Converting pfx (no password) to pem format..."
openssl pkcs12 -in certifier-nopass.pfx -out certifier-nopass.pem -nodes -password pass:""

echo "Converting pem to pfx with the password..."
openssl pkcs12 -export -out certifier.pfx -in certifier-nopass.pem -password pass:"$1"

echo "Converting pfx (with password) to pem..."
openssl pkcs12 -in certifier.pfx -nokeys -out certifier.pem -nodes -password pass:"$1"

echo "Extracting private key from pfx..."
openssl pkcs12 -in certifier.pfx -nocerts -out certifier-encrypted.key -passin pass:"$1" -passout pass:"$1"

echo "Decrypting private key..."
openssl rsa -in certifier-encrypted.key -out certifier.key -passin pass:"$1"

echo "*****************************************"

echo "******Acquiring Fluentd Certificate******"

echo "Downloading pfx (no password) from Key vault..."
secretId=$(az keyvault certificate show --vault-name $2 --name Fluentd | jq -r ".sid")
az keyvault secret download -f fluentd-nopass.pfx --vault-name $2 --id $secretId --encoding base64

echo "Converting pfx (no password) to pem format..."
openssl pkcs12 -in fluentd-nopass.pfx -out fluentd-nopass.pem -nodes -password pass:""

echo "Converting pem to pfx with the password..."
openssl pkcs12 -export -out fluentd.pfx -in fluentd-nopass.pem -password pass:"$1"

echo "Converting pfx (with password) to pem..."
openssl pkcs12 -in fluentd.pfx -nokeys -out fluentd.pem -nodes -password pass:"$1"

echo "Extracting private key from pfx..."
openssl pkcs12 -in fluentd.pfx -nocerts -out fluentd-encrypted.key -passin pass:"$1" -passout pass:"$1"

echo "Decrypting private key..."
openssl rsa -in fluentd-encrypted.key -out fluentd.key -passin pass:"$1"

echo "**************************************"

echo "******Acquiring Bootstrapper Key******"

echo "Downloading pfx (no password) from Key vault..."
secretId=$(az keyvault certificate show --vault-name $2 --name BootstrapperKey | jq -r ".sid")
az keyvault secret download -f bootstrapper-nopass.pfx --vault-name $2 --id $secretId --encoding base64

echo "Converting pfx (no password) to pem format..."
openssl pkcs12 -in bootstrapper-nopass.pfx -out bootstrapper-nopass.pem -nodes -password pass:""

echo "Converting pem to pfx with the password..."
openssl pkcs12 -export -out bootstrapper.pfx -in bootstrapper-nopass.pem -password pass:"$1"

echo "Converting pfx (with password) to pem..."
openssl pkcs12 -in bootstrapper.pfx -nokeys -out bootstrapper.pem -nodes -password pass:"$1"

echo "Extracting private key from pfx..."
openssl pkcs12 -in bootstrapper.pfx -nocerts -out bootstrapper-encrypted.key -passin pass:"$1" -passout pass:"$1"

echo "Decrypting private key..."
openssl rsa -in bootstrapper-encrypted.key -out bootstrapper.key -passin pass:"$1"

echo "*************************************"

# echo "******Acquiring Admin_operator Certificate******"

# echo "Downloading pfx (no password) from Key vault..."
# secretId=$(az keyvault certificate show --vault-name $2 --name admin-operator | jq -r ".sid")
# az keyvault secret download -f admin_operator-nopass.pfx --vault-name $2 --id $secretId --encoding base64

# echo "Converting pfx (no password) to pem format..."
# openssl pkcs12 -in admin_operator-nopass.pfx -out admin_operator-nopass.pem -nodes -password pass:""

# echo "Converting pem to pfx with the password..."
# openssl pkcs12 -export -out admin_operator.pfx -in admin_operator-nopass.pem -password pass:"$1"

# echo "Converting pfx (with password) to pem..."
# openssl pkcs12 -in admin_operator.pfx -nokeys -out admin_operator.pem -nodes -password pass:"$1"

# echo "Extracting private key from pfx..."
# openssl pkcs12 -in admin_operator.pfx -nocerts -out admin_operator-encrypted.key -passin pass:"$1" -passout pass:"$1"

# echo "Decrypting private key..."
# openssl rsa -in admin_operator-encrypted.key -out admin_operator.key.pem -passin pass:"$1"

# echo "**************************************"

echo "Removing temporary files..."
rm bootstrapper.pem
rm bootstrapper.pfx
rm bootstrapper-encrypted.key
rm bootstrapper-nopass.pem
rm bootstrapper-nopass.pfx
rm certifier.pfx
rm certifier-encrypted.key
rm certifier-nopass.pem
rm certifier-nopass.pfx
rm controller-encrypted.key
rm fluentd.pfx
rm fluentd-encrypted.key
rm fluentd-nopass.pem
rm fluentd-nopass.pfx
rm sonar-lte.pfx
rm sonar-nopass.pem
rm sonar-nopass.pfx
# rm admin_operator-nopass.pfx
# rm admin_operator-nopass.pem
# rm admin_operator-encrypted.key