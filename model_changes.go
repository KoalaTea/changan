package main

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/koalatea/changan/pkg/forms"
	"github.com/koalatea/changan/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

var (
	// ErrInvalid is an error for when an edited or added user is invalid
	ErrInvalid = errors.New("user is invalid")
)

func (app *App) addUser(user *forms.SignupUser) (*forms.SignupUser, error) {
	if !user.Valid() {
		return user, ErrInvalid
	}

	newUser := models.User{
		Username: user.Name,
		Password: user.Password,
		APIKey:   "not quite 8",
		Active:   false,
	}

	err := app.Database.AddUser(newUser)
	if err == models.ErrDuplicateEmail { // this error is not real TODO
		//add a form.Failures here TODO
		return nil, ErrInvalid //duplicate stuff
	} else if err != nil {
		return nil, err
	}

	return nil, nil
}

func (app *App) addDevice(newDevice *forms.NewDevice) (*forms.NewDevice, error) {
	var subnets []models.Subnet
	var err error
	gotSubnets := false
	if !newDevice.Valid() {
		return newDevice, ErrInvalid
	}

	// convert form.NewDevice.Interfaces and form.NewDevice.Interfaces.IPs to models.Device
	var interfaces []models.Interface
	for _, newDeviceInterface := range newDevice.Interfaces {
		var newInterface models.Interface
		newInterface.MAC = newDeviceInterface.MAC
		if newInterface.MAC == "" {
			newInterface.MAC = "FF:FF:FF:FF:FF:FF"
		}
		newInterface.Name = newDeviceInterface.Name
		if newInterface.Name == "" {
			newInterface.Name = "unknown" //TODO will need this changed when unique is added
		}
		var ips []models.IP
		for _, ip := range newDeviceInterface.IPs {
			newIP := models.IP{
				IP:       ip.IP,
				SubnetID: ip.SubnetID,
			}
			// If there is no subnet id we will figure out the subnet for the ip
			if newIP.SubnetID == bson.ObjectId("") {
				// If subnets have not been queried from the database yet get them
				if !gotSubnets {
					subnets, err = app.Mongo.GetAllSubnets()
					if err != nil {
						return nil, err
					}
					app.Logger.Debugf("api add device had no subnet id for ip '%s'", ip.IP)
					app.Logger.Debugf("so got subnets: %+v", subnets)
					gotSubnets = true
				}

				// go through all subnets and find out which one the ip belongs to
				var currentParent *models.Subnet
				var currentParentNet *net.IPNet
				netIP := net.ParseIP(ip.IP)
				for subnetLoc, subnet := range subnets {
					subnetIP, subnetIPNet, _ := net.ParseCIDR(fmt.Sprintf("%s/%d", subnet.IP, subnet.Mask))
					if subnetIPNet.Contains(netIP) {
						app.Logger.Debugf("device '%s' has ip '%s' that matched subnet %s", newDevice.Name,
							ip.IP, subnetIPNet)
						if newIP.SubnetID != bson.ObjectId("") {
							if currentParentNet.Contains(subnetIP) {
								app.Logger.Debugf("device '%s' ip '%s' had subnet '%s/%d' now has '%s/%d'",
									newDevice.Name, ip.IP, currentParent.IP, currentParent.Mask, subnet.IP,
									subnet.Mask)
								currentParent = &subnets[subnetLoc]
								currentParentNet = subnetIPNet
							}
						} else {
							app.Logger.Debugf("device '%s' ip '%s' had no parent now has '%s/%d'", newDevice.Name,
								ip.IP, subnet.IP, subnet.Mask)
							currentParent = &subnets[subnetLoc]
							currentParentNet = subnetIPNet
						}
					}
					// dynamically figure out ip
				}
				if currentParent == nil {
					id2 := bson.NewObjectId() // TODO when subnets are implemented make this auto figure out
					newIP.SubnetID = id2
				} else {
					newIP.SubnetID = currentParent.ID
				}
			}
			ips = append(ips, newIP)
		}
		newInterface.IPs = ips
		interfaces = append(interfaces, newInterface)
	}

	id := bson.NewObjectId()
	device := &models.Device{
		ID:         id,
		Name:       newDevice.Name,
		Team:       newDevice.Team,
		Owner:      newDevice.Owner,
		Location:   newDevice.Location,
		Interfaces: interfaces,
	}

	err = app.Mongo.AddDevice(device)
	return nil, err
}

func (app *App) addSubnet(subnet *forms.NewSubnet) (*forms.NewSubnet, error) {
	if !subnet.Valid() {
		return subnet, ErrInvalid
	}
	// verify that a default subnet exists
	app.InitializeServer()

	id := bson.NewObjectId()
	subnet.ID = id // verify id is not duplicate. because Insert subnet into tree assumes it is good
	newSubnet := &models.Subnet{
		ID:          id,
		Name:        subnet.Name,
		IP:          subnet.IP,
		Mask:        subnet.Mask,
		CIDR:        fmt.Sprintf("%s/%d", subnet.IP, subnet.Mask),
		HasChildren: false,
		CreatedTime: time.Now(),
		EditedTime:  time.Now(),
	}

	app.insertSubnetIntoTree(newSubnet)
	err := app.Mongo.AddSubnet(newSubnet)
	if err != nil {
		return nil, err
	}

	return subnet, nil
}

// TODO benchmark a duplicate error from mongo vs checking each id when looping through them here
func (app *App) insertSubnetIntoTree(newSubnet *models.Subnet) error {
	// May need to revisit this for optimization a secondary thought of approach was to instead do
	// a document to keep an active tree which would have a tree with names and id and also a list
	// of subnet IDs for checking if all subnets are in there (avoiding race conditions instead of
	// locking, though locking is a smarter idea I am unsure how to do it efficiently with a Database
	// would be nice if it can be done so that queries can still go through but then the race
	// condition may still exist ex. good for the page that is a tree, bad for anther insert subnet
	// that pulls the incomplete tree before it updates)
	//var subnets []models.Subnet
	var currentParent *models.Subnet
	var currentParentNet *net.IPNet
	subnetMap := make(map[string]models.Subnet)
	// test fix vet errors TODO figure out if I want this as my structure
	var err error

	app.Logger.Debugf("Adding subnet to tree for subnet id:%s name:%s subnet:%s/%d",
		newSubnet.ID.Hex(), newSubnet.Name, newSubnet.IP, newSubnet.Mask)
	subnets, err := app.Mongo.GetAllSubnets()
	if err != nil {
		return err
	}
	app.Logger.Debugf("Grabbed subnets:\n %+v", subnets)

	// go through the subnets to find the newSubnets parent
	newIP, newIPNet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", newSubnet.IP, newSubnet.Mask))
	if err != nil {
		return err
	}
	newSubnet.CIDR = newIPNet.String()
	for subnetLoc, subnet := range subnets {
		// keep subnets in a map for checking the parents children
		subnetMap[subnet.ID.Hex()] = subnet
		ip, ipNet, err := net.ParseCIDR(fmt.Sprintf("%s/%d", subnet.IP, subnet.Mask))
		if err != nil {
			return err
		}
		// TODO Look at this algorithm for any cases that would be wrong
		if currentParent != nil {
			app.Logger.Debugf("subnet %s current parent is cidr: %s/%d", newSubnet.ID.Hex(),
				currentParent.IP, currentParent.Mask)
		}
		// if this subnet contains the new subnet it could be the direct parent
		if ipNet.Contains(newIP) {
			app.Logger.Debugf("subnet '%s' has ip '%s' that matched subnet %s", newSubnet.Name,
				newIP, ipNet)
			// if the newSubnet already has a parent we will see if this new parent is closer to the
			// newSubnet than the currentParent, if there is no currentParent make this subnet the
			// currentParent
			if currentParent != nil {
				// if this subnet is within the currentParent, then it is closer to the newSubent and should
				// be the currentParent
				if currentParentNet.Contains(ip) {
					currentParent = &subnets[subnetLoc] //&subnet will create a pointer to subnet from range
					currentParentNet = ipNet
				}
			} else {
				currentParent = &subnets[subnetLoc]
				currentParentNet = ipNet
			}
		}
	}

	app.Logger.Debugf("Subnet's found parent is id: %s cidr: %s/%d", currentParent.ID.Hex(),
		currentParent.IP, currentParent.Mask)
	// go through the newSubnets parent's Children and decide if they belong to the newSubnet or
	// should belong to the parent still
	newParentChildren := []bson.ObjectId{newSubnet.ID}
	for _, child := range currentParent.Children {
		childSubnet := subnetMap[child.Hex()]
		ip, _, err := net.ParseCIDR(fmt.Sprintf("%s/%d", childSubnet.IP, childSubnet.Mask))
		if err != nil {
			return err
		}
		if newIPNet.Contains(ip) {
			newSubnet.Children = append(newSubnet.Children, childSubnet.ID)
			childSubnet.ParentID = newSubnet.ID // assuming the new subnet's ID is not a duplicate
			err = app.Mongo.EditSubnet(&childSubnet)
			if err != nil {
				return err
			}
		} else {
			newParentChildren = append(newParentChildren, childSubnet.ID)
		}
	}

	// edit the currentParent to have the new children
	if !currentParent.HasChildren {
		currentParent.HasChildren = true
	}
	currentParent.Children = newParentChildren
	err = app.Mongo.EditSubnet(currentParent)
	if err != nil {
		return err
	}

	if len(newSubnet.Children) != 0 {
		newSubnet.HasChildren = true
	}
	newSubnet.ParentID = currentParent.ID
	app.updateIPsSubnet(currentParent, newSubnet)
	return nil
}

func (app *App) updateIPsSubnet(oldSubnet *models.Subnet, newSubnet *models.Subnet) error {
	app.Logger.Debugf("updating ips from subnet '%s' to subnet '%s'", oldSubnet.CIDR, newSubnet.CIDR)
	devices, err := app.Mongo.GetFullDevicesForSubnet(oldSubnet)
	if err != nil {
		return err
	}
	_, subnetIPNet, err := net.ParseCIDR(newSubnet.CIDR)
	if err != nil {
		return err
	}
	for _, device := range devices {
		deviceChanged := false
		interfaces := []models.Interface{}
		for _, devInterface := range device.Interfaces {
			interfaceChanged := false
			ips := []models.IP{}
			for _, ip := range devInterface.IPs {
				if ip.SubnetID == oldSubnet.ID {
					netIP := net.ParseIP(ip.IP)
					if subnetIPNet.Contains(netIP) {
						app.Logger.Debugf("'%s' had subnet '%s' now has subnet '%s' with id '%s'", ip.IP,
							oldSubnet.CIDR, newSubnet.CIDR, newSubnet.ID.Hex())
						ip.SubnetID = newSubnet.ID
						deviceChanged = true
						interfaceChanged = true
						app.Logger.Debugf("the changed ip %+v", ip)
					}
					ips = append(ips, ip)
				}
			}
			if interfaceChanged {
				devInterface.IPs = ips
			}
			interfaces = append(interfaces, devInterface)
		}
		if deviceChanged {
			device.Interfaces = interfaces
			app.Logger.Debugf("device changed %+v", device)
			app.Mongo.EditDevice(&device)
		}
	}
	// TODO actually change these devices
	return nil
}
