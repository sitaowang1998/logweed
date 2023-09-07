import requests
import json

# URL to send the request to
url = "http://192.168.1.40:8080/clgsearch"

# JSON data to include in the request body
data = {
    "fid": "7,08e020d9a9",
    "archid": "fa9edcc7-c02a-4091-8245-f08947550b46",
    "nseg": "1",
    "container_1427088391284_0076_01_000267": "",
}

# Set the headers to indicate that the request body is JSON
headers = {
    "Content-Type": "application/json"
}

# Send the POST request with JSON data
response = requests.post(url, data=json.dumps(data), headers=headers)

# Print the response
print(response.status_code)
print(response.content)
