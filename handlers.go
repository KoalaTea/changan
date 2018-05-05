// Coder: koalatea
// Email: koalateac@gmail.com

package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/koalatea/changan/pkg/forms"
	"github.com/koalatea/changan/pkg/models"
	"gopkg.in/mgo.v2/bson"
)

func (app *App) Home(w http.ResponseWriter, r *http.Request) {
	app.RenderHTML(w, r, "home.page.html", nil)
}

func (app *App) viewDevices(w http.ResponseWriter, r *http.Request) {
	// This is one to one mapping multiple maps is there a better way? TODO
	// Map of deviceID to name of interface to mac address to a splice of ips
	// ex. {1: {'eth0': {'AA:BB:CC:DD:EE:FF': ['192.168.0.1', '192.168.0.2']}}}

	// Convert to Interfaces from the join table
	//devices, err := app.Mongo.GetAllDevices()
	devices, err := app.Mongo.GetAllDevices()
	if err != nil {
		app.ServerError(w, err)
		return
	}

	app.RenderHTML(w, r, "devices.page.html", &HTMLData{
		Devices: devices,
	})
}

func (app *App) SignupUser(w http.ResponseWriter, r *http.Request) {
	app.RenderHTML(w, r, "signup.page.html", &HTMLData{
		Form: &forms.SignupUser{},
	})
}

func (app *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ClientError(w, http.StatusBadRequest)
		return
	}

	form := &forms.SignupUser{
		Name:       r.PostForm.Get("name"),
		Password:   r.PostForm.Get("password"),
		RePassword: r.PostForm.Get("repassword"),
	}

	form, err = app.addUser(form)
	if err == ErrInvalid {
		app.RenderHTML(w, r, "signup.page.html", &HTMLData{Form: form})
		return
	} else if err != nil {
		app.ServerError(w, err)
		return
	}

	// TODO figure out messages might put them in HTMLData
	// Otherwise, add a confirmation flash message to the session confirming that
	// their signup worked and asking them to log in.
	msg := "Your signup was successful. Please log in using your credentials."
	session := app.Sessions.Load(r)
	err = session.PutString(w, "flash", msg)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (app *App) ReviewUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.Database.GetInactiveUsers()
	if err != nil {
		app.ServerError(w, err)
		return
	}

	app.RenderHTML(w, r, "review.users.page.html", &HTMLData{Users: users})
}

func (app *App) ViewUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	id, err := strconv.Atoi(userID)
	if err != nil {
		app.ServerError(w, err)
		return
	}
	user, err := app.Database.GetUser(id)
	if err != nil {
		app.ServerError(w, err)
		return
	}
	app.RenderHTML(w, r, "user.page.html", &HTMLData{User: user})
}

func (app *App) ActivateUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ClientError(w, http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.PostForm.Get("user_id"))
	if err != nil {
		app.ServerError(w, err)
		return
	}

	err = app.Database.SetActiveUser(id)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/user/review", http.StatusSeeOther)
}

func (app *App) LoginUser(w http.ResponseWriter, r *http.Request) {
	// flash message that signup is successful?
	app.RenderHTML(w, r, "login.page.html", &HTMLData{Form: &forms.LoginUser{}})
}

func (app *App) VerifyUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ClientError(w, http.StatusBadRequest)
		return
	}

	form := &forms.LoginUser{
		Username: r.PostForm.Get("username"),
		Password: r.PostForm.Get("password"),
	}

	if !form.Valid() {
		app.RenderHTML(w, r, "login.page.html", &HTMLData{Form: form})
		return
	}

	user := models.User{
		Username: form.Username,
		Password: form.Password,
	}

	currentUserID, err := app.Database.VerifyUser(user)
	//_, err = app.Database.VerifyUser(user)
	if err == models.ErrInvalidCredentials {
		form.Failures["Generic"] = "Email or Password is incorrect"
		app.RenderHTML(w, r, "login.page.html", &HTMLData{Form: form})
		return
	} else if err == models.ErrInactive {
		form.Failures["Generic"] = "Account is currently inactive"
		app.RenderHTML(w, r, "login.page.html", &HTMLData{Form: form})
		return
	} else if err != nil {
		app.ServerError(w, err)
		return
	}

	session := app.Sessions.Load(r)
	session.RenewToken(w)
	err = session.PutInt(w, "currentUserID", currentUserID)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	err = session.PutString(w, "Username", user.Username)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *App) LogoutUser(w http.ResponseWriter, r *http.Request) {
	session := app.Sessions.Load(r)
	err := session.Remove(w, "currentUserID")
	if err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/", 303)
}

func (app *App) viewDevice(w http.ResponseWriter, r *http.Request) {
	// TODO better this code return 404 if device is not id or does not exist
	// also check for positive numbers
	vars := mux.Vars(r)
	deviceID := vars["deviceID"]
	id := bson.ObjectIdHex(deviceID)

	searchDevice := &models.Device{ID: id}
	device, err := app.Mongo.GetDevice(searchDevice)
	if err != nil {
		app.ServerError(w, err)
		return
	}
	if device == nil { //TODO will it ever equal nil?
		app.NotFound(w)
		return
	}

	reports, err := app.Mongo.GetReportsForDevice(device)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	app.RenderHTML(w, r, "device.page.html", &HTMLData{
		Device:  device,
		Reports: reports,
	})
}

func (app *App) newDevice(w http.ResponseWriter, r *http.Request) {
	app.RenderHTML(w, r, "new.device.page.html", &HTMLData{
		Form: &forms.NewDevice{},
	})
}

func (app *App) createDevice(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ClientError(w, http.StatusBadRequest)
		return
	}

	form := &forms.NewDevice{
		Name:     r.PostForm.Get("name"),
		Team:     r.PostForm.Get("team"),
		Owner:    r.PostForm.Get("owner"),
		Location: r.PostForm.Get("location"),
	}

	if !form.Valid() {
		app.RenderHTML(w, r, "new.device.page.html", &HTMLData{Form: form})
		return
	}

	id := bson.NewObjectId()
	device := &models.Device{
		ID:       id,
		Name:     form.Name,
		Team:     form.Team,
		Owner:    form.Owner,
		Location: form.Location,
		//Interfaces: form.Interfaces,
	}

	err = app.Mongo.AddDevice(device) // TODO revisit if this should be pointer
	if err != nil {
		app.ServerError(w, err)
		return
	}
	/*
		device, err = app.Mongo.GetDeviceByName(*device)
		if err != nil {
			app.ServerError(w, err)
			return
		}*/

	http.Redirect(w, r, fmt.Sprintf("/devices/%s", id.Hex()), http.StatusSeeOther)
}

func (app *App) viewSubnets(w http.ResponseWriter, r *http.Request) {
	subnets, err := app.Mongo.GetAllSubnets()
	if err != nil {
		app.ServerError(w, err)
		return
	}

	app.RenderHTML(w, r, "subnets.page.html", &HTMLData{
		Subnets: subnets,
	})
}

func (app *App) viewSubnet(w http.ResponseWriter, r *http.Request) {
	// TODO better this code return 404 if device is not id or does not exist
	// also check for positive numbers
	vars := mux.Vars(r)
	subnetID := vars["subnetID"]
	id := bson.ObjectIdHex(subnetID)

	searchSubnet := &models.Subnet{ID: id}
	subnet, err := app.Mongo.GetSubnet(searchSubnet)
	if err != nil {
		app.ServerError(w, err)
		return
	}
	if subnet == nil {
		app.NotFound(w)
		return
	}

	devices, err := app.Mongo.GetDevicesForSubnet(subnet)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	app.RenderHTML(w, r, "subnet.page.html", &HTMLData{
		Subnet:  subnet,
		Devices: devices,
	})
}

func (app *App) createSubnet(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ClientError(w, http.StatusBadRequest)
		return
	}

	form := &forms.NewSubnet{
		Name:       r.PostForm.Get("name"),
		IP:         r.PostForm.Get("ip"),
		MaskString: r.PostForm.Get("mask"),
	}

	form, err = app.addSubnet(form)
	if err == ErrInvalid {
		app.RenderHTML(w, r, "new.subnet.page.html", &HTMLData{Form: form})
		return
	} else if err != nil {
		app.ServerError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/subnets/%s", form.ID.Hex()), http.StatusSeeOther)
}

func (app *App) viewReport(w http.ResponseWriter, r *http.Request) {
	// TODO better this code return 404 if device is not id or does not exist
	// also check for positive numbers
	vars := mux.Vars(r)
	reportID := vars["reportID"]
	id := bson.ObjectIdHex(reportID)

	searchReport := &models.Report{ID: id}
	report, err := app.Mongo.GetReport(searchReport)
	if err != nil {
		app.ServerError(w, err)
		return
	}
	if report == nil { //TODO will it ever equal nil?
		app.NotFound(w)
		return
	}

	app.RenderHTML(w, r, "report.page.html", &HTMLData{
		Report: report,
	})
}

func (app *App) newReport(w http.ResponseWriter, r *http.Request) {
	devices, err := app.Mongo.GetAllDevices()
	if err != nil {
		app.ServerError(w, err)
		return
	}
	app.RenderHTML(w, r, "new.report.page.html", &HTMLData{
		Form:    &forms.NewReport{},
		Devices: devices,
	})
}

func (app *App) createReport(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ClientError(w, http.StatusBadRequest)
		return
	}

	form := &forms.NewReport{
		DeviceID: bson.ObjectIdHex(r.PostForm.Get("device")),
		Title:    r.PostForm.Get("title"),
		Report:   r.PostForm.Get("report"),
	}

	if !form.Valid() {
		devices, err := app.Mongo.GetAllDevices()
		if err != nil {
			app.ServerError(w, err)
			return
		}
		app.RenderHTML(w, r, "new.report.page.html", &HTMLData{Form: form, Devices: devices})
		return
	}

	session := app.Sessions.Load(r)
	username, err := session.GetString("Username")
	if err != nil {
		app.ServerError(w, err)
		return
	}

	id := bson.NewObjectId()
	report := &models.Report{
		ID:       id,
		DeviceID: form.DeviceID,
		Title:    form.Title,
		Report:   form.Report,
		LastUser: username,
	}

	err = app.Mongo.AddReport(report)
	if err != nil {
		app.ServerError(w, err)
		return
	}

	// do devices and link to the device the report got added to?
	http.Redirect(w, r, fmt.Sprintf("/reports/%s", id.Hex()), http.StatusSeeOther)
}

/*
func viewIPs(w http.ResponseWriter, r *http.Request) {
	ips := GetAllIPs(BaseDB)

	data := &struct {
		Title string
		IPs   []IP
	}{
		Title: "ips",
		IPs:   ips,
	}

	t, _ := template.ParseFiles("templates/ips.html")
	if t == nil {
	}
	var buf bytes.Buffer

	err := t.Execute(&buf, data)
	if err != nil {
		panic(err)
	}
	_, err = buf.WriteTo(w)
	if err != nil {
		panic(err)
	}
}
*/
