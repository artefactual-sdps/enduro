import sys
from lxml import etree


def validate_xml(path, arelda):
    with open(path) as f:
        xml = f.read()
    doc = etree.fromstring(bytes(xml, encoding="UTF-8"))
    with open(arelda) as f:
        xsd_xml = f.read()
    xsd_doc = etree.fromstring(bytes(xsd_xml, encoding="UTF-8"))
    xsd = etree.XMLSchema(xsd_doc)
    valid = xsd.validate(doc)
    return valid


print("Is metadata.xml valid: ", validate_xml(sys.argv[1], sys.argv[2]))
