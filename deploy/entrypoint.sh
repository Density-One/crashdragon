#!/bin/sh
# entrypoint.sh

cat > /etc/crashdragon/config.toml <<EOF
[db]
connection = "host=${db_host} user=${db_user} dbname=${db_name} password=${db_password} sslmode=${db_sslmode}"
[directory]
assets = "/opt/crashdragon/share/crashdragon/assets"
content = "/opt/crashdragon/share/crashdragon/files"
templates = "/opt/crashdragon/share/crashdragon/templates"
[housekeeping]
reportretentiontime = "2190h"
[slack]
webhook = "${slack_webhook}"
[symbolicator]
executable = "./minidump_stackwalk"
trimmodulenames = true
[web]
bindaddress = ":8080"
bindsocket = "/var/run/crashdragon/crashdragon.sock"
usesocket = false
EOF

# Start application
exec "$@"
