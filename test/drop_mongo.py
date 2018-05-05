import pymongo

def main():
    client = pymongo.MongoClient()
    client.drop_database("changan_test")

if __name__ == '__main__':
    main()
