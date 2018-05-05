package forms

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/mgo.v2/bson"
)

// TODO why do this here? why not in database?

type NewDevice struct {
	ID         bson.ObjectId     `json:"device_id"`
	Name       string            `json:"device_name"`
	Team       string            `json:"team"`
	Owner      string            `json:"owner"`
	Location   string            `json:"location"`
	Interfaces []Interface       `json:"interfaces"`
	Failures   map[string]string // maybe change this?
}

type Interface struct {
	MAC  string `json:"mac"`
	Name string `json:"interface_name"`
	IPs  []IP   `json:"ips"`
}

type IP struct {
	SubnetID bson.ObjectId `json:"subnet_id"`
	IP       string        `json:"ip"`
}

type NewSubnet struct {
	ID         bson.ObjectId `json:"subnet_id"`
	Name       string        `json:"subnet_name"`
	IP         string        `json:"ip"`
	CIDR       string        `json:"cidr"`
	Mask       int           `json:"mask"`
	MaskString string
	Failures   map[string]string // TODO maybe change this?
}

type NewReport struct {
	ID       bson.ObjectId `json:"report_id"`
	DeviceID bson.ObjectId `json:"device_id"`
	Report   string        `json:"report"`
	Title    string        `json:"title"`
	Failures map[string]string
}

type SignupUser struct {
	Name       string
	Password   string
	RePassword string
	Failures   map[string]string
}

type LoginUser struct {
	Username string
	Password string
	Failures map[string]string
}

// TODO log invalid?
func (f *NewDevice) Valid() bool {
	f.Failures = make(map[string]string)

	// Check that the Title field is not blank and is not more than 100 characters
	// long. If it fails either of those checks, add a message to the f.Failures
	// map using the field name as the key
	if strings.TrimSpace(f.Name) == "" {
		f.Failures["Name"] = "Name is required"
	} else if utf8.RuneCountInString(f.Name) > 100 {
		f.Failures["Name"] = "Name can not be longer than 100 characters"
	}
	// Validate other fields here

	return len(f.Failures) == 0
}

func (f *NewSubnet) Valid() bool {
	f.Failures = make(map[string]string)

	// Check that the Title field is not blank and is not more than 100 characters
	// long. If it fails either of those checks, add a message to the f.Failures
	// map using the field name as the key
	if strings.TrimSpace(f.Name) == "" {
		f.Failures["Name"] = "Name is required"
	} else if utf8.RuneCountInString(f.Name) > 100 {
		f.Failures["Name"] = "Name can not be longer than 100 characters"
	}
	// Validate other fields here
	if strings.TrimSpace(f.IP) == "" {
		f.Failures["IP"] = "ip is required"
	} else if utf8.RuneCountInString(f.IP) > 45 {
		f.Failures["IP"] = "The largest ipv6 is 45 characters long"
	} // TODO validate ips are legitimate

	if f.MaskString != "" {
		mask, err := strconv.Atoi(f.MaskString)
		if err != nil {
			f.Failures["Mask"] = "Mask must be an integer"
		} else {
			f.Mask = mask
		}
	}

	if f.Mask > 128 {
		f.Failures["Mask"] = "The max ipv6 mask is 128"
	} // TODO other checks based on if it is ipv4 or ipv6

	return len(f.Failures) == 0
}

func (f *NewReport) Valid() bool {
	f.Failures = make(map[string]string)

	// Check that the Title field is not blank and is not more than 100 characters
	// long. If it fails either of those checks, add a message to the f.Failures
	// map using the field name as the key
	if strings.TrimSpace(f.Title) == "" {
		f.Failures["Title"] = "Title is required"
	} else if utf8.RuneCountInString(f.Title) > 100 {
		f.Failures["Title"] = "Title can not be longer than 100 characters"
	}
	// Validate other fields here
	if strings.TrimSpace(f.Report) == "" {
		f.Failures["Report"] = "Report is required"
	}

	return len(f.Failures) == 0
}

func (f *SignupUser) Valid() bool {
	f.Failures = make(map[string]string)

	if strings.TrimSpace(f.Name) == "" {
		f.Failures["Name"] = "Name is required"
	}

	if utf8.RuneCountInString(f.Password) < 8 {
		f.Failures["Password"] = "Password cannot be shorted than 8 characters"
	}
	if f.Password != f.RePassword {
		f.Failures["Password"] = "Passwords do not match"
		f.Failures["RePassword"] = "Passwords do not match"
	}

	return len(f.Failures) == 0
}

func (f *LoginUser) Valid() bool {
	f.Failures = make(map[string]string)

	if strings.TrimSpace(f.Username) == "" {
		f.Failures["Username"] = "Username is required"
	}

	if strings.TrimSpace(f.Password) == "" {
		f.Failures["Password"] = "Password is required"
	}

	return len(f.Failures) == 0
}
