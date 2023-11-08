import os
import sys
import subprocess
import bagit
from lxml import objectify
from datetime import date


def parseXML(xmlFile):
    """Parse the XML file"""
    with open(xmlFile) as f:
        xml = f.read()
    root = objectify.fromstring(bytes(xml, "UTF-8"))
    return root


def get_checksums(root):
    """Get checksums from metadata.xml"""
    for i in range(0, len(root.inhaltsverzeichnis.ordner)):
        dictionary = findchildren(root.inhaltsverzeichnis.ordner[i], path="")
    return dictionary


def findchildren(node, path, checksum="", dictionary={}):
    for el in node.getchildren():
        if el.tag == "{http://bar.admin.ch/arelda/v4}name":
            path = path + "/" + el.text
        if el.tag == "{http://bar.admin.ch/arelda/v4}pruefsumme":
            checksum = el.text
        findchildren(el, path)
        if checksum != "":
            dictionary[path] = checksum
    return dictionary


def find_checksum(file_path):
    command = "md5sum " + file_path
    checksum = subprocess.check_output(command, shell=True)
    return checksum


def writeMD5(files, path):
    """Write checksums.md5-file with checksums (md5) and file paths"""
    checksum_path = path + "/manifest-md5.txt"
    md5file = open(checksum_path, "a")
    metadata_checksum = find_checksum(path+"/data/header/metadata.xml")
    md5file.write(str(metadata_checksum)[2:34] + " data/header/metadata.xml \n")
    for filename, checksum in files.items():
        md5file.write(checksum + " data" + filename + "\n")
    md5file.close()


def create_checksums_file(bag_name):
    path = bag_name + "/data/header/metadata.xml"
    metadata_xml = parseXML(path)
    checksums = get_checksums(metadata_xml)
    writeMD5(checksums, bag_name)

# --- get SIP, unzip it and repackage it as a bagit bag ---


def repackage_sip(sip_name):
    bag_name = sip_name + "_bag"
    command = "mkdir " + bag_name + " && mkdir " + bag_name + "/data && cp -r " + sip_name + "/content " + sip_name + "/header " + bag_name + "/data && cp bagit.txt " + bag_name + "/."
    try:
        os.system(command)
    except:
        raise Exception("Could not make bag structure")
    try:
        create_checksums_file(bag_name)
    except:
        raise Exception("Checksum File could not be created")
    try:
        b = bagit.Bag(bag_name)
        if not b.is_valid():
            raise Exception("Bag not valid")
        b.info["Bagging-Date"] = date.strftime(date.today(), "%Y-%m-%d")
        b.info["Bag-Software-Agent"] = "bagit.py v%s <%s>" % (
                    b.version,
                    "https://github.com/LibraryOfCongress/bagit-python",
                )
        # Generates bag-info.txt with payload-oxum
        b.save(manifests=True)
    except:
        raise Exception("Something went wrong when repackaging the Bag")
    return (bag_name)


print(repackage_sip(sys.argv[1]))
