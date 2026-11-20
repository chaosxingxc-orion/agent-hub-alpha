"""Apply Docker's build proxy to Maven without retaining it in an image layer."""
import os
from pathlib import Path
from urllib.parse import unquote, urlsplit
from xml.etree.ElementTree import Element, SubElement, tostring

settings = Element("settings", xmlns="http://maven.apache.org/SETTINGS/1.2.0")
proxies = SubElement(settings, "proxies")
for protocol in ("http", "https"):
    value = os.environ.get(protocol.upper() + "_PROXY", "")
    if not value:
        continue
    url = urlsplit(value)
    if url.scheme not in ("http", "https") or not url.hostname:
        raise ValueError("Maven build requires an HTTP(S) proxy URL")
    proxy = SubElement(proxies, "proxy")
    values = {"id": "build-" + protocol, "active": "true", "protocol": protocol,
              "host": url.hostname, "port": str(url.port or 80),
              "nonProxyHosts": "localhost|127.0.0.1"}
    if url.username:
        values["username"] = unquote(url.username)
    if url.password:
        values["password"] = unquote(url.password)
    for key, item in values.items():
        SubElement(proxy, key).text = item
target = Path("/tmp/maven-settings.xml")
target.write_bytes(tostring(settings, encoding="utf-8", xml_declaration=True))
target.chmod(0o600)
