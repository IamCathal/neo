import requests
import json

URL = "http://localhost:2590/api/insertgame"

#  https://api.steampowered.com/ISteamApps/GetAppList/v2/
file = open("appIDs.json")
data = json.load(file)
count = 0

for i in data["applist"]["apps"]:
    newObj = {
        "appid": i["appid"],
        "name": i["name"]
    }
    req = requests.post(URL, json=newObj)
    print(f"[{count}] {req.text}")
    count += 1

