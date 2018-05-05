package models

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Device struct {
	ID         bson.ObjectId `bson:"_id" json:"device_id"`
	Name       string        `bson:"device_name" json:"device_name"`
	Team       string        `bson:"team" json:"team"`
	Owner      string        `bson:"owner" json:"owner"`
	Location   string        `bson:"location" json:"location"`
	Interfaces []Interface   `bson:"interfaces" json:"interfaces"`
}

type Interface struct {
	//ID   int    `bson:"_id"`
	IPs  []IP   `bson:"ips" json:"ips"`
	MAC  string `bson:"mac" json:"mac"`
	Name string `bson:"interface_name" json:"interface_name"`
}

type IP struct {
	SubnetID bson.ObjectId `bson:"subnet_id" json:"subnet_id"`
	IP       string        `bson:"ip" json:"ip"`
}

type Subnet struct {
	ID          bson.ObjectId   `bson:"_id" json:"subnet_id"`
	Name        string          `bson:"subnet_name" json:"subnet_name"`
	IP          string          `bson:"ip" json:"ip"`
	Mask        int             `bson:"mask" json:"mask"`
	CIDR        string          `bson:"cidr" json:"cidr"`
	ParentID    bson.ObjectId   `bson:"parent_id,omitempty" json:"-"`
	HasChildren bool            `bson:"has_children" json:"-"`
	Children    []bson.ObjectId `bson:"children" json:"-"`
	CreatedTime time.Time       `bson:"created_timestamp" json:"created_timestamp"`
	EditedTime  time.Time       `bson:"edited_timestamp" json:"edited_timestamp"`
}

type Report struct {
	ID       bson.ObjectId `bson:"_id" json:"report_id"`
	DeviceID bson.ObjectId `bson:"device_id" json:"device_id"`
	Title    string        `bson:"title" json:"title"`
	Report   string        `bson:"report" json:"report"`
	LastUser string        `bson:"user"`
}

type User struct {
	UserID   int    `db:"user_id"`
	GroupID  int    `db:"group_id"`
	Username string `db:"username"`
	Password string `db:"password"`
	APIKey   string `db:"api_key"`
	Active   bool   `db:"active"`
}
