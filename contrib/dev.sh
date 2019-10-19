#!/bin/sh

export BASIC_AUTH="true"
export AUTH_URL="http://basic-auth-plugin:8080/validate"

# Secrets should be created even if basic-auth is disabled.
echo "Attempting to create credentials for gateway.."
echo "admin" | docker secret create basic-auth-user -
echo "localdev" | docker secret create basic-auth-password -


docker stack deploy func --compose-file docker-compose.yml


printf '%-15s:\t %s\n' 'Username' 'admin'
printf '%-15s:\t %s\n' 'Password' 'localdev'
printf '%-15s:\t %s\n' 'CLI Auth' 'echo -n localdev | faas-cli login --username=admin --password-stdin'
