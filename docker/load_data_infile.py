from zipfile import ZipFile, ZipExtFile
import json
import os
import MySQLdb

config = {
    'host': os.environ.get('MYSQL_HOST', 'localhost'),
    'port': int(os.environ.get('MYSQL_PORT', '3306')),
    'user': os.environ.get('MYSQL_USER', 'root'),
    'passwd': os.environ.get('MYSQL_PASSWORD', ''),
    'db': os.environ.get('MYSQL_DB', ''),
    'charset': 'utf8mb4',
    'cursorclass': MySQLdb.cursors.DictCursor,
}

accountTableNames = ["id", "fname", "sname", "phone", "sex", "birth", "country", "city", "joined", "status", "premium_start", "premium_end"]
interestTableNames = ["account_id", "interest"]
likeTableNames = ["account_id_from", "account_id_to", "ts"]

def extractFromJson(j):
    accounts = []
    interests = []
    likes = []

    for account in j["accounts"]:
        sa = {account.get(i) for i in accountTableNames}
        sa["sex"] = ["f", "m"].index(sa["sex"])
        sa["status"] = ["свободны", "заняты", "всё сложно"].index(sa["status"])
        accounts.append(sa)

        it = [{"account_id": sa["id"], "interest" : interest} for interest in account["interests"]]
        interests.append(it)

        lk = [{"account_id_from": sa["id"], "account_id_to": ll["id"], "ts" : ll["ts"]} for ll in account["likes"]]
        likes.append(lk)
    return accounts, interests, likes


def importData(file: ZipExtFile):
    print(f"extracting ${file.name} ...")
    j = json.load(file)
    accounts, interests, likes = extractFromJson(j)



with ZipFile('data.zip', 'r') as archive:
    for name in archive.namelist():
        with archive.open(name) as file:
            importData(file)
