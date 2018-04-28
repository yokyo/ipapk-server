package api

import (
	"bytes"
	"fmt"
	"image/png"
	"net/http"
	"path/filepath"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
	"github.com/gin-gonic/gin"
	"github.com/phinexdaz/ipapk"
	"github.com/phinexdaz/ipapk-server/conf"
	"github.com/phinexdaz/ipapk-server/models"
	"github.com/phinexdaz/ipapk-server/serializers"
	"github.com/phinexdaz/ipapk-server/utils"
	"github.com/teris-io/shortid"
)

func Upload(c *gin.Context) {
	changelog := c.PostForm("changelog")
	file, err := c.FormFile("file")
	if err != nil {
		fmt.Println("ERROR: FormFile")
		return
	}

	ext := models.BundleFileExtension(filepath.Ext(file.Filename))
	if !ext.IsValid() {
		fmt.Println("ERROR: BundleFileExtension")
		return
	}

	uuid, err := shortid.Generate()
	if err != nil {
		return
	}

	filename := utils.GetAppPath(uuid + string(ext.PlatformType().Extention()))

	if err := c.SaveUploadedFile(file, filename); err != nil {
		fmt.Println("ERROR: SaveUploadedFile - ")
		return
	}

	app, err := ipapk.NewAppParser(filename)
	if err != nil {
		fmt.Println("ERROR: NewAppParser")
		return
	}

	if err := utils.SaveIcon(app.Icon, uuid+".png"); err != nil {
		fmt.Println("ERROR: SaveIcon")
		return
	}

	bundle := new(models.Bundle)
	bundle.UUID = uuid
	bundle.PlatformType = ext.PlatformType()
	bundle.Name = app.Name
	bundle.BundleId = app.BundleId
	bundle.Version = app.Version
	bundle.Build = app.Build
	bundle.Size = app.Size
	bundle.ChangeLog = changelog

	if err := models.AddBundle(bundle); err != nil {
		fmt.Println("ERROR: AddBundle")
		return
	}

	c.JSON(http.StatusOK, &serializers.BundleJSON{
		UUID:       uuid,
		Name:       bundle.Name,
		Platform:   bundle.PlatformType.String(),
		BundleId:   bundle.BundleId,
		Version:    bundle.Version,
		Build:      bundle.Build,
		InstallUrl: bundle.GetInstallUrl(conf.AppConfig.ProxyURL()),
		QRCodeUrl:  conf.AppConfig.ProxyURL() + bundle.GetQrCode(),
		IconUrl:    conf.AppConfig.ProxyURL() + bundle.GetIcon(),
		Changelog:  bundle.ChangeLog,
		Downloads:  bundle.Downloads,
	})
}

func GetQRCode(c *gin.Context) {
	uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(uuid)
	if err != nil {
		return
	}

	data := fmt.Sprintf("%v/bundle/%v?_t=%v", conf.AppConfig.ProxyURL(), bundle.UUID, time.Now().Unix())
	code, err := qr.Encode(data, qr.L, qr.Unicode)
	if err != nil {
		return
	}
	code, err = barcode.Scale(code, 160, 160)
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, code); err != nil {
		return
	}

	c.Data(http.StatusOK, "image/png", buf.Bytes())
}

func GetChangelog(c *gin.Context) {
	uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(uuid)
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "change.html", gin.H{
		"changelog": bundle.ChangeLog,
	})
}

func GetBundle(c *gin.Context) {
	uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(uuid)
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"bundle":     bundle,
		"installUrl": bundle.GetInstallUrl(conf.AppConfig.ProxyURL()),
		"qrCodeUrl":  conf.AppConfig.ProxyURL() + bundle.GetQrCode(),
		"iconUrl":    conf.AppConfig.ProxyURL() + bundle.GetIcon(),
	})
}

func GetVersions(c *gin.Context) {
	uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(uuid)
	if err != nil {
		return
	}

	versions, err := bundle.GetVersions()
	if err != nil {
		return
	}

	c.HTML(http.StatusOK, "version.html", gin.H{
		"versions": versions,
		"uuid":     bundle.UUID,
	})
}

func GetBuilds(c *gin.Context) {
	uuid := c.Param("uuid")
	version := c.Param("version")

	bundle, err := models.GetBundleByUID(uuid)
	if err != nil {
		return
	}

	builds, err := bundle.GetBuilds(version)
	if err != nil {
		return
	}

	var bundles []serializers.BundleJSON
	for _, v := range builds {
		bundles = append(bundles, serializers.BundleJSON{
			UUID:       v.UUID,
			Name:       v.Name,
			Platform:   v.PlatformType.String(),
			BundleId:   v.BundleId,
			Version:    v.Version,
			Build:      v.Build,
			InstallUrl: v.GetInstallUrl(conf.AppConfig.ProxyURL()),
			QRCodeUrl:  conf.AppConfig.ProxyURL() + v.GetQrCode(),
			IconUrl:    conf.AppConfig.ProxyURL() + v.GetIcon(),
			Changelog:  v.ChangeLog,
			Downloads:  v.Downloads,
		})
	}

	c.HTML(http.StatusOK, "build.html", gin.H{
		"builds": bundles,
	})
}

func GetPlist(c *gin.Context) {
	uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(uuid)
	if err != nil {
		return
	}

	if bundle.PlatformType != models.BundlePlatformTypeIOS {
		return
	}

	ipaUrl := conf.AppConfig.ProxyURL() + "/bundle/" + bundle.UUID + "/download"

	data, err := models.NewPlist(bundle.Name, bundle.Version, bundle.BundleId, ipaUrl).Marshall()
	if err != nil {
		return
	}

	c.Data(http.StatusOK, "application/x-plist", data)
}

func DownloadAPP(c *gin.Context) {
	uuid := c.Param("uuid")

	bundle, err := models.GetBundleByUID(uuid)
	if err != nil {
		return
	}

	go bundle.UpdateDownload()

	downloadUrl := conf.AppConfig.ProxyURL() + bundle.GetApp()
	c.Redirect(http.StatusFound, downloadUrl)
}
