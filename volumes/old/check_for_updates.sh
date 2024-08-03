#!/bin/bash

# Check if wget is installed, if not, install it using Homebrew
command -v wget >/dev/null 2>&1 || { echo >&2 "I require wget but it's not installed. Installing now..."; brew install wget; }

pwd=$(pwd)

# Define variables
url="https://repo1.maven.org/maven2/org/mongodb/kafka/mongo-kafka-connect/"
html=$(curl -s $url)
download_dir=$pwd/volumes
env_file="$pwd/.env" # Change this to the path of your .env file

latest_version=$(echo "$html" \
| awk '/<a href="[^"]*"/ {split($0,a,"\""); print a[2]}' \
| ggrep -oP '^\K[1-9]\d*(\.\d+)*\/?$' \
| sort -V \
| tail -n 1 \
| sed 's/\/$//')

echo "MongoDB Kafka connector latest Version: $latest_version"

# Construct the download link
download_link="${url}${latest_version}/mongo-kafka-connect-${latest_version}-all.jar"


# Check if the file already exists
if [ ! -f $download_dir/mongo-kafka-connect-${latest_version}-all.jar ]; then
    # If the file does not exist, download it
    wget -Ps $download_dir "$download_link"
else
    echo "File already exists."
fi

if [ $? -ne 0 ]; then
    echo "Failed to download the JAR file."
    exit 1
fi

# Read the current version from the .env file
current_version=$(grep MONGO_KAFKA_CONNECT_VERSION $env_file | awk -F '=' '{print $2}')

# Check if the variable was found and update it
if [ ! -z "$current_version" ]; then
    # If the variable exists, update it
    echo "Updating MONGO_KAFKA_CONNECT_VERSION from $current_version to $latest_version"
    sed -i "s/${current_version}/${latest_version}/g" "${env_file}"
fi

echo "Latest version of mongo-kafka-connect has been downloaded and updated in the .env file."
