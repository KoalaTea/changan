package models

import (
	"errors"
	"log"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Database struct {
	*mgo.Database
}

func (db Database) GetAllDevices() ([]Device, error) {
	var devices []Device
	err := db.C("devices").Find(bson.M{}).All(&devices)
	return devices, err
}

func (db Database) GetDevice(device *Device) (*Device, error) {
	err := db.C("devices").FindId(device.ID).One(device)
	return device, err
}

func (db Database) GetDeviceByName(name string) (*Device, error) {
	device := &Device{}
	err := db.C("devices").Find(bson.M{"device_name": name}).One(device)
	return device, err
}

func (db Database) GetDeviceByIP(ip string) (*Device, error) {
	device := &Device{}
	err := db.C("devices").Find(bson.M{"interfaces.ips.ip": ip}).One(device)
	return device, err
}

func (db Database) AddDevice(device *Device) error {
	err := db.C("devices").Insert(device)
	return err
}

func (db Database) DeleteDevice(device *Device) error {
	err := db.C("devices").Remove(device) // By id?
	return err
}

func (db Database) EditDevice(device *Device) error {
	err := db.C("devices").UpdateId(device.ID, device)
	return err
}

func (db Database) GetDevicesForSubnet(subnet *Subnet) ([]Device, error) {
	var devices []Device
	err := db.C("devices").Find(bson.M{"interfaces.ips.subnet_id": subnet.ID}).
		Select(bson.M{"_id": 1, "device_name": 1, "interfaces.ips.ip": 1,
			"interfaces.ips.subnet_id": 1}).All(&devices)
	return devices, err
}

// TODO better way of dealing with Mongo
func (db Database) GetFullDevicesForSubnet(subnet *Subnet) ([]Device, error) {
	var devices []Device
	err := db.C("devices").Find(bson.M{"interfaces.ips.subnet_id": subnet.ID}).All(&devices)
	return devices, err
}

func (db Database) GetAllSubnets() ([]Subnet, error) {
	var subnets []Subnet
	err := db.C("subnets").Find(bson.M{}).All(&subnets)
	return subnets, err
}

func (db Database) GetSubnet(subnet *Subnet) (*Subnet, error) {
	err := db.C("subnets").FindId(subnet.ID).One(subnet)
	return subnet, err
}

func (db Database) GetSubnetByName(name string) (*Subnet, error) {
	subnet := &Subnet{}
	err := db.C("subnets").Find(bson.M{"subnet_name": name}).One(subnet)
	return subnet, err
}

func (db Database) AddSubnet(subnet *Subnet) error {
	err := db.C("subnets").Insert(subnet)
	return err
}

func (db Database) DeleteSubnet(subnet *Subnet) error {
	err := db.C("subnets").Remove(subnet) // By id?
	return err
}

func (db Database) EditSubnet(subnet *Subnet) error {
	err := db.C("subnets").UpdateId(subnet.ID, subnet)
	return err
}

func (db Database) GetAllReports() ([]Report, error) {
	var reports []Report
	err := db.C("reports").Find(bson.M{}).All(&reports)
	return reports, err
}

func (db Database) GetReport(report *Report) (*Report, error) {
	err := db.C("reports").FindId(report.ID).One(report)
	return report, err
}

func (db Database) AddReport(report *Report) error {
	err := db.C("reports").Insert(report)
	return err
}

func (db Database) GetReportsForDevice(device *Device) ([]Report, error) {
	var reports []Report
	err := db.C("reports").Find(bson.M{"device_id": device.ID}).All(&reports)
	return reports, err
}

func OpenMongo(dsn string) (*mgo.Session, *mgo.Database, error) {
	var db *mgo.Database
	session, err := mgo.Dial("127.0.0.1")
	if err != nil {
		return nil, nil, err
	}

	db = session.DB(dsn)
	return session, db, nil
}

// Declare a Database type (for now it's just an empty struct).
type SQLDatabase struct {
	*sqlx.DB
}

// This error does not exist in my project TODO
var (
	ErrDuplicateEmail     = errors.New("models: email address already in use")
	ErrInvalidCredentials = errors.New("models: invalid user credentials")
	ErrInactive           = errors.New("models: account is inactive")
)

func (db SQLDatabase) AddUser(user User) error {
	// move this out of here or maybe not
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, password, api_key, active) VALUES(?, ?, ?, ?)",
		user.Username, string(hashedPassword), user.APIKey, user.Active)
	if err != nil {
		if err.(*mysql.MySQLError).Number == 1062 {
			return ErrDuplicateEmail // TODO fix this error the error is not real in my project
		}
	}
	return err
}

func (db SQLDatabase) VerifyUser(user User) (int, error) {
	var foundUser User
	err := db.Get(&foundUser,
		"SELECT user_id, username, password, active FROM users where username=?", user.Username)
	if err != nil {
		//return 0, err // use if you have no idea what is happening
		return 0, ErrInvalidCredentials
	}

	if !foundUser.Active {
		return 0, ErrInactive
	}
	// might move this elsewhere
	err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(user.Password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, ErrInvalidCredentials
	} else if err != nil {
		return 0, err
	}

	return foundUser.UserID, nil
}

func (db SQLDatabase) GetInactiveUsers() ([]User, error) {
	var users []User
	err := db.Select(&users,
		"SELECT user_id, username, password, active FROM users WHERE active=false")
	if err != nil {
		return users, err
	}

	return users, nil
}

func (db SQLDatabase) SetActiveUser(id int) error {
	_, err := db.Exec("UPDATE users SET active=true WHERE user_id=?", id)
	return err
}

func (db SQLDatabase) GetUser(id int) (*User, error) {
	user := &User{}
	err := db.Get(user, "SELECT user_id, username, password, active FROM users WHERE user_id=?", id)
	return user, err
}

func OpenMysqlDB(dsn string) *sqlx.DB {
	var db *sqlx.DB
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	return db
}
