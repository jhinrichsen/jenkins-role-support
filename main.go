package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const (
	GlobalRole  = "globalRoles"
	ProjectRole = "projectRoles"
	SlaveRole   = "slaveRoles"
)

// Only allow projectRoles to be changed
type Roles struct {
	Roles []Role `json:"roles"`
}

type Role struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	Pattern     string   `json:"pattern"`
	Users       []string `json:"users"`
}

func load(filename string) (Roles, error) {
	var roles Roles
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return roles, err
	}
	err = json.Unmarshal(buf, &roles)
	if err != nil {
		return roles, err
	}
	return roles, nil
}

// ServerInstance holds remote REST services
type ServerInstance struct {
	Protocol, Server   string
	Port               int
	Context            string
	Username, Password string
}

// BaseURL returns the base URL for a remote endpoint
func (a ServerInstance) BaseURL() string {
	s := fmt.Sprintf("%s://%s:%d", a.Protocol, a.Server, a.Port)
	if len(a.Context) > 0 {
		s += a.Context
	}
	return s
}

// SetBasicAuth does exactly that
func (a ServerInstance) SetBasicAuth(req *http.Request) {
	req.SetBasicAuth(a.Username, a.Password)
}

// JenkinsInstance holds access information to a remote Jenkins server
type JenkinsInstance struct {
	ServerInstance
}

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// AssignRole correlates a user with a role
func (a JenkinsInstance) AssignRole(roleType, name, sid string) {

	p := fmt.Sprintf("/role-strategy/strategy/assignRole")
	req, err := http.NewRequest(http.MethodPost, a.BaseURL()+p, nil)
	die(err)
	a.SetBasicAuth(req)

	// Handle parameters
	q := req.URL.Query()
	q.Add("type", roleType)
	q.Add("roleName", name)
	q.Add("sid", sid)
	req.URL.RawQuery = q.Encode()

	log.Printf("posting %+v", req)
	res, err := (&http.Client{}).Do(req)
	die(err)
	if res.StatusCode != 200 {
		log.Fatalf("Expected status code 200 but got %d\n",
			res.StatusCode)
	}
	log.Printf("assigned user %s to role %s\n", name, sid)
}

// AddRole expects permissions to be Jenkins class names such as
// jenkins.item.Build
// Jenkins will create the role, silently skipping non-existent permissions
// pattern is optional, not used for global roles
func (a JenkinsInstance) AddRole(roleType string, name string,
	permissions string, pattern string, overwrite bool) {

	p := fmt.Sprintf("/role-strategy/strategy/addRole")
	req, err := http.NewRequest(http.MethodPost, a.BaseURL()+p, nil)
	die(err)
	a.SetBasicAuth(req)

	// Handle parameters
	q := req.URL.Query()
	q.Add("type", roleType)
	q.Add("roleName", name)
	if len(pattern) > 0 {
		// Jenkins' default is .*
		q.Add("pattern", pattern)
	}
	q.Add("permissionIds", permissions)
	q.Add("overwrite", strconv.FormatBool(overwrite))
	req.URL.RawQuery = q.Encode()

	log.Printf("posting %+v", req)
	res, err := (&http.Client{}).Do(req)
	die(err)
	if res.StatusCode != 200 {
		log.Fatalf("Expected status code 200 but got %d\n",
			res.StatusCode)
	}
	log.Printf("added role %s\n", name)
}

// Roles lists all available roles, as of Role Strategy Plugin 2.6.1 only
// for type=globalRoles
func (a JenkinsInstance) Roles(roleType string) {
	p := fmt.Sprintf("/role-strategy/strategy/getAllRoles?type=%s",
		roleType)
	req, err := http.NewRequest(http.MethodGet, a.BaseURL()+p, nil)
	die(err)
	a.SetBasicAuth(req)

	log.Printf("posting %+v", req)
	res, err := (&http.Client{}).Do(req)
	die(err)
	if res.StatusCode != 200 {
		log.Fatalf("Expected status code 200 but got %d\n",
			res.StatusCode)
	}
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)
	die(err)
	log.Printf("roles %+v\n", string(buf))
}

func main() {
	protocol := flag.String("protocol", "http", "Jenkins protocol")
	hostname := flag.String("hostname", "localhost", "Jenkins hostname")
	port := flag.Int("port", 8080, "Jenkins REST port")
	context := flag.String("context", "", "Jenkins REST context")
	username := flag.String("username", "admin", "Jenkins REST auth")
	password := flag.String("password", "admin", "Jenkins REST auth")

	action := flag.String("action", "getAllRoles", "REST action")
	overwrite := flag.Bool("overwrite", false,
		"Allow overwriting role")
	pattern := flag.String("pattern", ".*", "Role plugin pattern")
	permissions := flag.String("permissions",
		"hudson.model.Item.Discover,hudson.model.Item.Build",
		"Comma separated list of Jenkins permissions")
	roleName := flag.String("name", "testrole1", "Role name")
	roleType := flag.String("type", "globalRoles",
		"Role Strategy Plugin role type")
	sid := flag.String("sid", "user1", "Jenkins SID")
	flag.Parse()

	j := JenkinsInstance{ServerInstance{*protocol, *hostname, *port,
		*context, *username, *password}}
	switch *action {
	case "getAllRoles":
		j.Roles(*roleType)
	case "assignRole":
		j.AssignRole(*roleType, *roleName, *sid)
		j.Roles(*roleType)
	case "addRole":
		j.AddRole(*roleType, *roleName, *permissions, *pattern,
			*overwrite)
		j.Roles(*roleType)
	}
}
