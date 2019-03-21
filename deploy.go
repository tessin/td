package td

import (
	"encoding/base64"
	"fmt"
	"log"
	"path/filepath"

	"github.com/tessin/td/az"
)

func Deploy(resourceGroup, name string) error {
	ctx, err := az.GetContext(az.ResourceManager)
	if err != nil {
		return err
	}

	xml, err := ctx.GetWebPublishXml(resourceGroup, name)
	if err != nil {
		return err
	}

	log.Println(xml)

	kudu := &KuduClient{
		Base:          fmt.Sprintf("https://%v.scm.azurewebsites.net", name),
		Authorization: "Basic" + " " + base64.StdEncoding.EncodeToString([]byte(xml.User+":"+xml.Pass)),
	}

	abs, _ := filepath.Abs(".")
	zip := filepath.Join(abs, filepath.Base(abs)+".zip")

	err = kudu.ZipDeploy(zip)
	if err != nil {
		return err
	}

	log.Println("npm install --production")

	out, err := kudu.Command("npm install --production", "/home/site/wwwroot")
	if err != nil {
		return err
	}

	log.Println(out.Output)

	log.Println("done")

	return nil
}
