#!/bin/bash

# Change to the directory wtih our code that we plan to work from.
cd "$GOPATH/src/coverd"

# Delete any local binaries (not directories) so they wont be uploaded.
echo "==== Releasing GoBlog ===="
echo " Deleting the local binary if it exists (so it isn't uploaded)..."
rm coverd
echo " Done"

# Delete any files on prod sever from previous build.
echo " Deleting existing code..."
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "rm -rf /root/go/src/GoBlog"
echo " Code deleted successfully!"

# Upload source code to prod via rsync (file transfer and sync across multi servers).
echo " Uploading code..."
# The \ at the end of the line tells bash that our
# command isn't done and wraps to the next line.
rsync -avr --exclude '.git/*' --exclude 'tmp/*' \
--exclude 'images/*' ./ \
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1":/root/go/src/GoBlog/
echo " Code uploaded successfully!"

# Go get third-party libraries for the application.
echo " Go getting deps..."
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "export GOPATH=/root/go; \
/usr/local/go/bin/go get golang.org/x/crypto/bcrypt"
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "export GOPATH=/root/go; \
/usr/local/go/bin/go get github.com/gorilla/mux"
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "export GOPATH=/root/go; \
/usr/local/go/bin/go get github.com/gorilla/schema"
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "export GOPATH=/root/go; \
/usr/local/go/bin/go get github.com/lib/pq"
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "export GOPATH=/root/go; \
/usr/local/go/bin/go get github.com/jinzhu/gorm"
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "export GOPATH=/root/go; \
/usr/local/go/bin/go get github.com/gorilla/csrf"

# Build application. Use -o flag to define the name of the binary.
# Build from /root/app directory so our binary is created there.
echo " Building the code on the remote server "
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" 'export GOPATH=/root/go; \
cd /root/app; \
/usr/local/go/bin/go build -o ./server \
$GOPATH/src/GoBlog/*.go'
echo " Code built successfully!"

# Move assets, views, and Caddyfile intp the /root/app directory.
echo " Moving assets..."
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "cd /root/app; \
cp -R /root/go/src/GoBlog/assets ."
echo " Assets moved successfully!"
echo " Moving views..."
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "cd /root/app; \
cp -R /root/go/src/GoBlog/views ."
echo " Views moved successfully!"
echo " Moving Caddyfile..."
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "cd /root/app; \
cp /root/go/src/GoBlog/Caddyfile ."
echo " Views moved successfully!"

# Restart services created previously, 
# restart Caddy service (in case we pushed changes to Caddyfile)
echo " Restarting the server..."
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "sudo service GoBlog restart"
echo " Server restarted successfully!"
echo " Restarting Caddy server..."
gcloud compute --project "goblog-2018" ssh --zone "us-east1-c" "goblog-instance-1" "sudo service caddy restart"
echo " Caddy restarted successfully!"
echo "==== Done releasing Coverd ===="