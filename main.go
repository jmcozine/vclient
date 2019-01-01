package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
)

func getEnvString(v string, def string) string {
	r := os.Getenv(v)
	if r == "" {
		return def
	}

	return r
}

const envURL = "VCSA_URL"

var urlFlag = flag.String("url", getEnvString(envURL, "https://vcsa.example.com"+vim25.Path), fmt.Sprintf("ESXi or vCenter address [%s]", envURL))

type foo struct {
	User   string
	Client *govmomi.Client
	VMs    []mo.VirtualMachine
	Auth   bool
	Token  string
	Host   string
}

func main() {
	flag.Parse()

	u, err := soap.ParseURL(*urlFlag)
	if err != nil {
		log.Fatal("ParseURL: ", err)
	}

	tmpl := template.Must(template.ParseFiles("template.html"))
	ctx := context.Background()
	account := foo{}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")
			account.User = username

			if username != "" && password != "" {
				u.User = url.UserPassword(username, password)
			}

			c, err := govmomi.NewClient(ctx, u, true)
			if err != nil {
				log.Fatal("NewClient: ", err)
			}
			defer c.Logout(ctx)
			account.Client = c

			if c.IsVC() {
				account.Auth = true

				m := view.NewManager(c.Client)
				v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
				if err != nil {
					log.Fatal("CreateContainerView: ", err)
				}
				defer v.Destroy(ctx)

				var vms []mo.VirtualMachine
				err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
				if err != nil {
					log.Fatal("Retrieve: ", err)
				}
				log.Println("Got list of vms")
				log.Println(vms[0].Summary.Config.Name)
				account.VMs = vms

				tmpl.Execute(w, account)
			}
		} else if r.FormValue("vm") != "" {
			ok, err := account.Client.SessionManager.SessionIsActive(ctx)
			if err != nil {
				log.Panic(err)
			}
			if !ok {
				log.Println("Logging in with username/password")
				account.Client.SessionManager.Login(ctx, u.User)
				defer account.Client.SessionManager.Logout(ctx)
			}
			f := find.NewFinder(account.Client.Client, true)
			dc, err := f.DefaultDatacenter(ctx)
			if err != nil {
				log.Fatal("NewFinder: ", err)
			}
			f.SetDatacenter(dc)

			vm, err := f.VirtualMachine(ctx, r.FormValue("vm"))
			if err != nil {
				log.Fatal(err)
			}

			ticket, err := vm.AcquireTicket(ctx, "webmks")
			if err != nil {
				log.Fatal("AcquireTicket: ", err)
			}
			account.Token = ticket.Ticket
			account.Host = ticket.Host

			tmpl.Execute(w, account)
		} else if account.Client == nil || !account.Client.IsVC() {
			tmpl.Execute(w, nil)
			return
		}
	})

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static", fs))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
