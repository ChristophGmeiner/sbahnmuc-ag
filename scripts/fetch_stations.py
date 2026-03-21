import urllib.request
import urllib.parse
import xml.etree.ElementTree as ET
from time import sleep

stations = [
"München Karlsplatz", "München-Allach", "Baierbrunn", "Buchenau (Oberbay)", "München-Daglfing",
"Deisenhofen", "München Donnersbergerbrücke", "Ebenhausen-Schäftlarn", "Erding", "Feldafing",
"Feldkirchen (b München)", "München-Feldmoching", "München Flughafen Terminal", "München-Freiham", "Freising",
"Fürstenfeldbruck", "Geltendorf", "Germering-Unterpfaffenhofen", "München-Giesing", "Gilching-Argelsried",
"Grafrath", "Großhesselohe Isartalbf", "Haar", "München Hackerbrücke", "München Harras",
"München Heimeranplatz", "Herrsching", "Höhenkirchen-Siegertsbrunn", "Höllriegelskreuth", "Icking"
]

client_id = "d3a73ffba9c2ca165575f54dea0c678f"
api_key   = "e153cbbf91fc748d9dd3360c1ff50dad"

evas = []

for s in stations:
    search_name = urllib.parse.quote(s)
    api_url = f"https://apis.deutschebahn.com/db-api-marketplace/apis/timetables/v1/station/{search_name}"
    
    db_req = urllib.request.Request(api_url, headers={
        "DB-Client-Id": client_id,
        "DB-Api-Key": api_key,
        "Accept": "application/xml"
    })
    
    try:
        resp = urllib.request.urlopen(db_req)
        xml_data = resp.read()
        root = ET.fromstring(xml_data)
        
        # Take the first exact/similar match
        best_eva = None
        best_name = None
        for sn in root.findall('station'):
            if sn.get('eva'):
                best_eva = sn.get('eva')
                best_name = sn.get('name')
                break
                
        if best_eva:
            evas.append(f'\t\t"{best_eva}", // {best_name}')
            print(f"Resolved: {best_name} ({best_eva})")
        else:
            print(f"Not found: {s}")
        sleep(0.1)
    except Exception as e:
        print(f"Error {s}: {e}")

print("\nGo Slice for cmd/local/main.go:")
print("stations := []string{")
for e in evas:
    print(e)
print("\t}")
