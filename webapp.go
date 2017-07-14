package main

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"git.1750studios.com/GSoC/CrashDragon/config"
	"git.1750studios.com/GSoC/CrashDragon/database"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	uuid "github.com/satori/go.uuid"
)

// PostCrashComment allows you to post a comment to a crash
func PostCrashComment(c *gin.Context) {
	User := c.MustGet("user").(database.User)
	var Crash database.Crash
	database.Db.First(&Crash, "id = ?", c.Param("id"))
	if Crash.ID == uuid.Nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var Comment database.Comment
	database.Db.FirstOrInit(&Comment)
	Comment.UserID = User.ID
	Comment.ID = uuid.NewV4()
	unsafe := blackfriday.MarkdownCommon([]byte(c.PostForm("comment")))
	Comment.Content = template.HTML(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
	if len(strings.TrimSpace(string(Comment.Content))) == 0 {
		c.Redirect(http.StatusMovedPermanently, "/crashes/"+Crash.ID.String())
		return
	}
	Comment.CrashreportID = uuid.Nil
	Comment.CrashID = Crash.ID
	database.Db.Save(&Comment)
	c.Redirect(http.StatusMovedPermanently, "/crashes/"+Crash.ID.String()+"#comment-"+Comment.ID.String())
}

// PostCrashreportComment allows you to post a comment to a crashreport
func PostCrashreportComment(c *gin.Context) {
	User := c.MustGet("user").(database.User)
	var Crashreport database.Crashreport
	database.Db.First(&Crashreport, "id = ?", c.Param("id"))
	if Crashreport.ID == uuid.Nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	var Comment database.Comment
	database.Db.FirstOrInit(&Comment)
	Comment.UserID = User.ID
	Comment.ID = uuid.NewV4()
	unsafe := blackfriday.MarkdownCommon([]byte(c.PostForm("comment")))
	Comment.Content = template.HTML(bluemonday.UGCPolicy().SanitizeBytes(unsafe))
	if len(strings.TrimSpace(string(Comment.Content))) == 0 {
		c.Redirect(http.StatusMovedPermanently, "/crashreports/"+Crashreport.ID.String())
	}
	Comment.CrashreportID = Crashreport.ID
	Comment.CrashID = uuid.Nil
	database.Db.Save(&Comment)
	c.Redirect(http.StatusMovedPermanently, "/crashreports/"+Crashreport.ID.String()+"#comment-"+Comment.ID.String())
}

// GetCrashes returns crashes
func GetCrashes(c *gin.Context) {
	var Crashes []database.Crash
	sort := c.DefaultQuery("sort", "all_crash_count")
	switch sort {
	case "all_crash_count":
		sort = "all_crash_count"
	case "win_crash_count":
		sort = "win_crash_count"
	case "mac_crash_count":
		sort = "mac_crash_count"
	case "lin_crash_count":
		sort = "lin_crash_count"
	case "first_reported":
		sort = "first_reported"
	case "last_reported":
		sort = "last_reported"
	default:
		sort = "all_crash_count"
	}
	order := c.DefaultQuery("order", "desc")
	switch order {
	case "desc":
		order = "DESC"
	case "asc":
		order = "ASC"
	default:
		order = "DESC"
	}
	database.Db.Order(sort + " " + order).Find(&Crashes)
	c.HTML(http.StatusOK, "crashes.html", gin.H{
		"title": "Crashes",
		"items": Crashes,
	})
}

// GetCrash returns details of a crash
func GetCrash(c *gin.Context) {
	var Crash database.Crash
	var Crashreports []database.Crashreport
	var Comments []database.Comment
	database.Db.First(&Crash, "id = ?", c.Param("id")).Order("created_at DESC").Related(&Crashreports).Order("created_at DESC").Related(&Comments)
	for i, Comment := range Comments {
		database.Db.Model(&Comment).Related(&Comments[i].User)
	}
	c.HTML(http.StatusOK, "crash.html", gin.H{
		"title":    "Crash",
		"items":    Crashreports,
		"comments": Comments,
		"ID":       Crash.ID.String(),
	})
}

// GetCrashreports returns crashreports
func GetCrashreports(c *gin.Context) {
	var Reports []database.Crashreport
	var List []struct {
		ID        string
		Signature string
		Date      time.Time
		Product   string
		Version   string
		Platform  string
		Reason    string
		Location  string
	}
	sort := c.DefaultQuery("sort", "created_at")
	switch sort {
	case "product":
		sort = "product"
	case "version":
		sort = "version"
	case "os":
		sort = "os"
	default:
		sort = "created_at"
	}
	order := c.DefaultQuery("order", "desc")
	switch order {
	case "desc":
		order = "DESC"
	case "asc":
		order = "ASC"
	default:
		order = "DESC"
	}
	database.Db.Where("processed = true").Order(sort + " " + order).Find(&Reports)
	for _, Report := range Reports {
		var Item struct {
			ID        string
			Signature string
			Date      time.Time
			Product   string
			Version   string
			Platform  string
			Reason    string
			Location  string
		}
		Item.ID = Report.ID.String()
		Item.Date = Report.CreatedAt
		Item.Product = Report.Product
		Item.Version = Report.Version
		Item.Platform = Report.Os
		Item.Reason = Report.Report.CrashInfo.Type
		for _, Frame := range Report.Report.CrashingThread.Frames {
			if Frame.File == "" && Item.Signature != "" {
				continue
			}
			Item.Signature = Frame.Function
			if Frame.File == "" {
				continue
			}
			Item.Location = path.Base(Frame.File) + ":" + strconv.Itoa(Frame.Line)
			break
		}
		List = append(List, Item)
	}
	c.HTML(http.StatusOK, "crashreports.html", gin.H{
		"title": "Crashreports",
		"items": List,
	})
}

// GetCrashreport returns details of crashreport
func GetCrashreport(c *gin.Context) {
	var Report database.Crashreport
	var Comments []database.Comment
	database.Db.First(&Report, "id = ?", c.Param("id")).Order("created_at DESC").Related(&Comments)
	for i, Comment := range Comments {
		database.Db.Model(&Comment).Related(&Comments[i].User)
	}
	var Item struct {
		ID        string
		Signature string
		Date      time.Time
		Product   string
		Version   string
		Platform  string
		Arch      string
		Processor string
		Reason    string
		Location  string
		Comment   string
		Uptime    string
	}
	Item.ID = Report.ID.String()
	Item.Date = Report.CreatedAt
	Item.Product = Report.Product
	Item.Version = Report.Version
	Item.Platform = Report.Os + " " + Report.OsVersion
	Item.Arch = Report.Arch
	Item.Processor = Report.Report.SystemInfo.CPUInfo + " (" + strconv.Itoa(Report.Report.SystemInfo.CPUCount) + " cores)"
	Item.Reason = Report.Report.CrashInfo.Type
	Item.Comment = Report.Comment
	h := (Report.ProcessUptime / 3600000) % 24
	m := (Report.ProcessUptime / 60000) % 60
	s := (Report.ProcessUptime / 1000) % 60
	Item.Uptime = fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	for _, Frame := range Report.Report.CrashingThread.Frames {
		if Frame.File == "" && Item.Signature != "" {
			continue
		}
		Item.Signature = Frame.Function
		if Frame.File == "" {
			continue
		}
		Item.Location = path.Base(Frame.File) + ":" + strconv.Itoa(Frame.Line)
		break
	}
	result, _ := c.Cookie("result")
	if result != "" {
		c.SetCookie("result", "", 1, "/", "", false, false)
	}
	c.HTML(http.StatusOK, "crashreport.html", gin.H{
		"title":    "Crashreport",
		"item":     Item,
		"report":   Report.Report,
		"result":   result,
		"comments": Comments,
	})
}

// GetCrashreportFile returns minidump file of crashreport
func GetCrashreportFile(c *gin.Context) {
	var Crashreport database.Crashreport
	if err := database.Db.Where("id = ?", c.Param("id")).First(&Crashreport).Error; err != nil {
		c.AbortWithError(http.StatusNotFound, err)
		return
	}
	name := c.Param("name")
	switch name {
	case "upload_file_minidump":
		file := path.Join(config.C.ContentDirectory, "Crashreports", Crashreport.ID.String()[0:2], Crashreport.ID.String()[0:4], Crashreport.ID.String()+".dmp")
		f, err := os.Open(file)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		c.Header("Content-Disposition", "attachment; filename=\""+Crashreport.ID.String()+".dmp\"")
		c.Data(http.StatusOK, "application/octet-stream", data)
		return
	case "processed_json":
		c.Header("Content-Disposition", "attachment; filename=\""+Crashreport.ID.String()+".json\"")
		c.Data(http.StatusOK, "application/json", []byte(Crashreport.ReportContentJSON))
		return
	case "processed_txt":
		c.Header("Content-Disposition", "attachment; filename=\""+Crashreport.ID.String()+".txt\"")
		c.Data(http.StatusOK, "text/plain", []byte(Crashreport.ReportContentTXT))
		return
	default:
		c.AbortWithError(http.StatusBadRequest, errors.New(name+" is a unknwon file"))
		return
	}
}

// GetSymfiles returns symfiles
func GetSymfiles(c *gin.Context) {
	var Symfiles []database.Symfile
	database.Db.Find(&Symfiles)
	c.HTML(http.StatusOK, "symfiles.html", gin.H{
		"title": "Symfiles",
		"items": Symfiles,
	})
}

// GetSymfile returns content of symfile
func GetSymfile(c *gin.Context) {
	var Symfile database.Symfile
	database.Db.Where("id = ?", c.Param("id")).First(&Symfile)
	f, err := os.Open(path.Join(config.C.ContentDirectory, "Symfiles", Symfile.Name, Symfile.Code, Symfile.Name+".sym"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  err.Error,
		})
		return
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": http.StatusBadRequest,
			"error":  err.Error,
		})
		return
	}
	c.Data(http.StatusOK, "text/plain", data)
}
