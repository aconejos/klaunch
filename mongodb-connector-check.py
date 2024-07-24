import requests
import os
import re
from dotenv import load_dotenv, set_key

def get_latest_version():
    url = "https://search.maven.org/solrsearch/select?q=g:org.mongodb.kafka+AND+a:mongo-kafka-connect&rows=1&wt=json"
    response = requests.get(url)
    data = response.json()
    return data['response']['docs'][0]['latestVersion']

def download_jar(version):
    url = f"https://repo1.maven.org/maven2/org/mongodb/kafka/mongo-kafka-connect/{version}/mongo-kafka-connect-{version}.jar"
    response = requests.get(url)
    if response.status_code == 200:
        with open(f"mongo-kafka-connect-{version}.jar", "wb") as f:
            f.write(response.content)
        print(f"Successfully downloaded mongo-kafka-connect-{version}.jar")
        return True
    else:
        print(f"Failed to download version {version}")
        return False

def update_env_file(version):
    dotenv_path = '.env'
    load_dotenv(dotenv_path)
    set_key(dotenv_path, "MONGO_KAFKA_CONNECT_VERSION", version)
    print(f"Updated .env file with MONGO_KAFKA_CONNECT_VERSION={version}")

def main():
    version = input("Enter the version number (x.y.z) or press Enter for the latest version: ")
    
    if not version:
        version = get_latest_version()
        print(f"Using latest version: {version}")
    
    if not re.match(r'^\d+\.\d+\.\d+$', version):
        print("Invalid version format. Please use x.y.z format.")
        return

    if download_jar(version):
        update_env_file(version)

if __name__ == "__main__":
    main()
