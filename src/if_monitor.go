package main

import (
	"errors"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

type Update interface {
	getLinkId() string
	handleUpdate() (netlink.Link, error)
}

type LinkUpdate struct {
	Action   string
	LinkId   string
	Object   string
	Command  string
	Argument string
	link     netlink.Link
}

func UpdateKernel(updateChan chan Update, resyncC <-chan time.Time) {
	log.Info("Interface monitoring thread started.")

	for {
		select {
		case update := <-updateChan:
			log.WithField("update", update).Debug("Link update")
			link, err := update.handleUpdate();
			if err != nil {
				//todo handle fail.retry or alert
			}
			UpdateEtcd(update.getLinkId(), link)

		case <-resyncC: //periodic resyncs
			log.Debug("Resync trigger")
			err := resync()
			if err != nil {
				log.WithError(err).Fatal("Failed to read link states from netlink.")
			}
		}
	}
	log.Fatal("Failed to read events from Netlink.")
}
func resync() error {
	return nil
}

func UpdateMap(id string, updatedLink netlink.Link) {
}

func UpdateEtcd(id string, updatedLink netlink.Link) {
}

func (update LinkUpdate) getLinkId() string {
	return update.LinkId
}

func (update LinkUpdate) handleUpdate() (netlink.Link, error) {
	link := update.link
	updateError := errors.New("update fail, " + update.Command + update.Argument + link.Attrs().Name)
	switch update.Action {
	case "update":
		if update.Command == "set" && update.Argument == "up" {
			if err := netlink.LinkSetUp(link); err != nil {
				log.Error("update fail", err)
				return nil, updateError
			}
		}
		if update.Command == "set" && update.Argument == "down" {
			if err := netlink.LinkSetDown(link); err != nil {
				log.Error("update fail", err)
				return nil, updateError
			}
		}
		updatedLink, _ := GetLinkByName(link.Attrs().Name) // link will not be update,you should retrieve the link by your self.should i do this here or return the updated link?
		log.WithFields(log.Fields{
			"link name":   updatedLink.Attrs().Name,
			"link status": updatedLink.Attrs().Flags,
		}).Debug("set ", update.Argument, " linux success")
		return updatedLink, nil
	case "del":
	//return
	case "add":
	//return
	}
	return nil, errors.New("no action found")
}
