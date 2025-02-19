package controllers

import (
	"LibrarySystemGolang/models"
	"LibrarySystemGolang/utils"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

func ToLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}
func Logout(c *gin.Context) {
	c.SetCookie("session", "", -1, "/", "", false, true)
	c.SetCookie("loginStatus", "", -1, "/", "", false, true)
	c.SetCookie("isAdmin", "", -1, "/", "", false, true)    // 清除管理员角色标识
	c.SetCookie("readercard", "", -1, "/", "", false, true) // 清除读者 ID
	c.Redirect(http.StatusFound, "/login")
}

func LoginCheck(c *gin.Context) {
	id := c.PostForm("id")
	passwd := c.PostForm("passwd")
	isAdmin := hasMatchAdmin(id, passwd)
	isReader := hasMatchReader(id, passwd)

	res := map[string]string{}
	if isAdmin {
		// 管理员登录
		admin := getAdmin(id)
		c.SetCookie("admin", strconv.FormatInt(admin.AdminID, 10), 3600, "/", "", false, true)
		c.SetCookie("user_name", admin.Username, 3600, "/", "", false, true)
		c.SetCookie("loginStatus", "true", 3600, "/", "", false, true)
		c.SetCookie("isAdmin", "true", 3600, "/", "", false, true) // 设置管理员角色标识
		res["stateCode"] = "1"
		res["msg"] = "管理员登陆成功！"
	} else if isReader {
		// 读者登录
		readerCard := getReaderCard(id)
		c.SetCookie("readercard", strconv.FormatInt(readerCard.ReaderID, 10), 3600, "/", "", false, true)
		c.SetCookie("user_name", readerCard.Username, 3600, "/", "", false, true)
		c.SetCookie("loginStatus", "true", 3600, "/", "", false, true)
		c.SetCookie("isAdmin", "", -1, "/", "", false, true) // 清除管理员角色标识
		res["stateCode"] = "2"
		res["msg"] = "读者登陆成功！"
	} else {
		res["stateCode"] = "0"
		res["msg"] = "账号或密码错误！"
	}
	c.JSON(http.StatusOK, res)
}
func GetCurrentUserInfo(c *gin.Context) {
	userName, _ := c.Cookie("user_name")
	c.JSON(http.StatusOK, gin.H{
		"username": userName,
	})
}
func AdminMain(c *gin.Context) {
	cookie, err := c.Cookie("admin")
	if err != nil {
		log.Println("Error fetching admin cookie:", err)
		c.Redirect(http.StatusFound, "/login")
		return
	}
	adminID, err := strconv.ParseInt(cookie, 10, 64)
	if err != nil {
		log.Println("Error parsing admin ID from cookie:", err)
		c.Redirect(http.StatusFound, "/login")
		return
	}
	admin := getAdmin(strconv.FormatInt(adminID, 10))
	if admin.AdminID == 0 {
		log.Println("Admin not found")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "admin_main.html", gin.H{
		"Username": admin.Username,
	})
}

func ReaderMain(c *gin.Context) {
	cookie, err := c.Cookie("readercard")
	if err != nil {
		log.Println("Error fetching reader card cookie:", err)
		c.Redirect(http.StatusFound, "/login")
		return
	}
	readerID, err := strconv.ParseInt(cookie, 10, 64)
	if err != nil {
		log.Println("Error parsing reader ID from cookie:", err)
		c.Redirect(http.StatusFound, "/login")
		return
	}
	readerCard := getReaderCard(strconv.FormatInt(readerID, 10))
	if readerCard.ReaderID == 0 {
		log.Println("Reader not found")
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.HTML(http.StatusOK, "reader_main.html", gin.H{
		"Username": readerCard.Username,
	})
}

func AdminShowRePassWrodPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_repasswd.html", nil)
}

func AdminRePassWrod(c *gin.Context) {
	adminIDStr, _ := c.Cookie("admin")
	var req struct {
		OldPasswd string `json:"oldPasswd"`
		NewPasswd string `json:"newPasswd"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	inputOldPasswd := req.OldPasswd
	newPasswd := req.NewPasswd

	adminId, _ := strconv.ParseInt(adminIDStr, 10, 64)
	dbOldPasswd := getAdminPassword(adminId)
	if dbOldPasswd != inputOldPasswd {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不正确！"})
		return
	}
	if dbOldPasswd == newPasswd {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码与旧密码一致"})
		return
	}

	if adminRePassword(adminId, newPasswd) {
		c.SetCookie("loginStatus", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"success": "密码修改成功！"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": "密码修改失败！"})
	}
}

func ReaderShowRePassWordPage(c *gin.Context) {
	c.HTML(http.StatusOK, "reader_repasswd.html", nil)
}

func ReaderRePassWord(c *gin.Context) {
	readerIdStr, _ := c.Cookie("readercard")

	var req struct {
		OldPasswd string `json:"oldPasswd"`
		NewPasswd string `json:"newPasswd"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	inputOldPasswd := req.OldPasswd
	newPasswd := req.NewPasswd

	readerId, _ := strconv.ParseInt(readerIdStr, 10, 64)
	dbOldPasswd := getReaderPassword(readerId)
	if dbOldPasswd != inputOldPasswd {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不正确！"})
		return
	}
	if dbOldPasswd == newPasswd {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码与旧密码一致"})
		return
	}
	if readerRePassword(readerId, newPasswd) {
		c.SetCookie("loginStatus", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"success": "密码修改成功！"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": "密码修改失败！"})
	}
}

func NotFound(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", nil)
}

// 数据库操作函数
func hasMatchAdmin(id string, password string) bool {
	var count int64
	if err := utils.DB.Model(&models.Admin{}).Where("admin_id = ? AND password = ?", id, password).Count(&count).Error; err != nil {
		log.Println("Error checking admin:", err)
		return false
	}
	return count == 1
}

func hasMatchReader(id string, password string) bool {
	var count int64
	if err := utils.DB.Model(&models.ReaderCard{}).Where("reader_id = ? AND password = ?", id, password).Count(&count).Error; err != nil {
		log.Println("Error checking reader:", err)
		return false
	}
	return count > 0
}

func getAdmin(id string) models.Admin {
	var admin models.Admin
	if err := utils.DB.Where("admin_id = ?", id).First(&admin).Error; err != nil {
		log.Println("Error fetching admin:", err)
		return models.Admin{}
	}
	return admin
}

func getReaderCard(id string) models.ReaderCard {
	var readerCard models.ReaderCard
	if err := utils.DB.Where("reader_id = ?", id).First(&readerCard).Error; err != nil {
		log.Println("Error fetching reader card:", err)
		return models.ReaderCard{}
	}
	return readerCard
}

func getAdminPassword(id int64) string {
	var password string
	if err := utils.DB.Model(&models.Admin{}).Where("admin_id = ?", id).Select("password").First(&password).Error; err != nil {
		log.Println("Error fetching admin password:", err)
		return ""
	}
	return password
}

func getReaderPassword(id int64) string {
	var password string
	if err := utils.DB.Model(&models.ReaderCard{}).Where("reader_id = ?", id).Select("password").First(&password).Error; err != nil {
		log.Println("Error fetching reader password:", err)
		return ""
	}
	return password
}

func adminRePassword(id int64, newPassword string) bool {
	if err := utils.DB.Model(&models.Admin{}).Where("admin_id = ?", id).Update("password", newPassword).Error; err != nil {
		log.Println("Error updating admin password:", err)
		return false
	}
	return true
}

func readerRePassword(id int64, newPassword string) bool {
	if err := utils.DB.Model(&models.ReaderCard{}).Where("reader_id = ?", id).Update("password", newPassword).Error; err != nil {
		log.Println("Error updating reader password:", err)
		return false
	}
	return true
}
