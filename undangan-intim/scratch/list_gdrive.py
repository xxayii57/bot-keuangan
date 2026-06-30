import urllib.request
import re
import ssl

ssl._create_default_https_context = ssl._create_unverified_context

url = "https://drive.google.com/drive/folders/1HhLO7BPtKy4w4mN5hKGDDPvAKsw8T__Q"
req = urllib.request.Request(
    url,
    headers={'User-Agent': 'Mozilla/5.0'}
)

try:
    with urllib.request.urlopen(req) as response:
        html = response.read().decode('utf-8')
    
    # Extract file names from Google Drive initial data state JSON
    matches = re.findall(r'\["([^"]+\.(?:mp4|mov|mp3|jpg|png|jpeg))"', html)
    print("Found files:", list(set(matches)))
except Exception as e:
    print("Error:", e)
