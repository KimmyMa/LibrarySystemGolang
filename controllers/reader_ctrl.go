package controllers

import (
	"LibrarySystemGolang/models"
	"LibrarySystemGolang/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// AdminReader 获取所有读者信息
func AdminReader(c *gin.Context) {
	var readers []models.ReaderInfo
	if err := utils.DB.Find(&readers).Error; err != nil {
		fmt.Println("Error fetching readers:", err)
		c.HTML(http.StatusOK, "admin_reader.html", gin.H{"error": "无法获取读者信息"})
		return
	}
	c.HTML(http.StatusOK, "admin_reader.html", gin.H{"readers": readers})
}

// AdminReaderDelete 删除读者信息
func AdminReaderDelete(c *gin.Context) {
	readerIDStr := c.Param("id")
	readerID, err := strconv.ParseInt(readerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的读者ID"})
		return
	}

	// 删除读者卡信息
	if err := utils.DB.Delete(&models.ReaderCard{}, readerID).Error; err != nil {
		fmt.Println("Error deleting reader card:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除读者卡信息失败"})
		return
	}
	// 删除读者信息
	if err := utils.DB.Delete(&models.ReaderInfo{}, readerID).Error; err != nil {
		fmt.Println("Error deleting reader info:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除读者信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "删除成功"})
}

// ReaderInfo 显示读者信息页面
func ReaderInfo(c *gin.Context) {
	readerIDStr, err := c.Cookie("readercard")
	if err != nil {
		c.HTML(http.StatusOK, "reader_info.html", gin.H{"error": "未找到读者信息"})
		return
	}
	readerID, _ := strconv.ParseInt(readerIDStr, 10, 64)

	readerInfo := getReaderInfo(readerID)
	if readerInfo == nil {
		c.HTML(http.StatusOK, "reader_info.html", gin.H{"error": "读者信息未找到"})
		return
	}
	c.HTML(http.StatusOK, "reader_info.html", gin.H{"readerinfo": readerInfo})
}

func ReaderShowInfoUpdatePage(c *gin.Context) {
	readerIDStr := c.Param("id")

	readerID, _ := strconv.ParseInt(readerIDStr, 10, 64)

	readerInfo := getReaderInfo(readerID)
	if readerInfo == nil {
		c.HTML(http.StatusOK, "reader_info_update.html", gin.H{"error": "读者信息未找到"})
		return
	}
	c.HTML(http.StatusOK, "reader_info_update.html", gin.H{"readerInfo": readerInfo})
}

// ReaderInfo 显示读者信息页面
func ReaderInfoUpdate(c *gin.Context) {
	readerIDStr := c.Param("id")
	readerID, err := strconv.ParseInt(readerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无效的读者ID"})
		return
	}

	var readerInfo models.ReaderInfo
	if err := c.ShouldBindJSON(&readerInfo); err != nil {
		fmt.Println("Error binding data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "绑定数据失败"})
		return
	}

	readerInfo.ReaderID = readerID
	if err := utils.DB.Where("reader_id = ?", readerID).Save(&readerInfo).Error; err != nil {
		fmt.Println("Error editing reader info:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读者信息修改失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "读者信息修改成功"})
}

// AdminReaderQuery 显示编辑读者信息页面
func AdminReaderQuery(c *gin.Context) {
	readerIDStr := c.Param("id")
	readerID, err := strconv.ParseInt(readerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的读者ID"})
		return
	}

	readerInfo := getReaderInfo(readerID)
	if readerInfo == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "读者信息未找到"})
		return
	}
	c.HTML(http.StatusOK, "admin_reader_update.html", gin.H{"readerInfo": readerInfo})
}

// AdminReaderUpdate 处理编辑读者信息操作
func AdminReaderUpdate(c *gin.Context) {
	readerIDStr := c.Param("id")
	readerID, err := strconv.ParseInt(readerIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无效的读者ID"})
		return
	}

	var readerInfo models.ReaderInfo
	if err := c.ShouldBindJSON(&readerInfo); err != nil {
		fmt.Println("Error binding data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "绑定数据失败"})
		return
	}

	readerInfo.ReaderID = readerID
	if err := utils.DB.Where("reader_id = ?", readerID).Save(&readerInfo).Error; err != nil {
		fmt.Println("Error editing reader info:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读者信息修改失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "读者信息修改成功"})
}

// AdminShowReaderAddPage 显示添加读者信息页面
func AdminShowReaderAddPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_reader_add.html", nil)
}

// AdminReaderCreate 处理添加读者信息操作
func AdminReaderCreate(c *gin.Context) {

	type FrontData struct {
		PassWord string           `json:"password"`
		Name     string           `json:"name"`
		Sex      string           `json:"sex"`
		Birth    models.LocalDate `json:"birth"`
		Address  string           `json:"address"`
		Phone    string           `json:"phone"`
	}

	var frontData FrontData
	if err := c.ShouldBindJSON(&frontData); err != nil {
		fmt.Println("Error binding data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "绑定数据失败"})
		return
	}
	var readerInfo models.ReaderInfo
	readerInfo.Name = frontData.Name
	readerInfo.Sex = frontData.Sex
	readerInfo.Birth = frontData.Birth
	readerInfo.Address = frontData.Address
	readerInfo.Phone = frontData.Phone
	if err := utils.DB.Create(&readerInfo).Error; err != nil {
		fmt.Println("Error adding reader info:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加读者信息失败"})
		return
	}
	frontData.PassWord, _ = generateSaltedHash(frontData.PassWord)
	// 添加读者卡信息
	readerCard := models.ReaderCard{
		ReaderID: readerInfo.ReaderID,
		Username: readerInfo.Name,
		Password: frontData.PassWord, // 使用前端发送的密码
	}
	if err := utils.DB.Create(&readerCard).Error; err != nil {
		fmt.Println("Error adding reader card:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "添加读者卡信息失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "添加读者信息成功, 读者卡号为: " + strconv.FormatInt(readerInfo.ReaderID, 10)})
}

// 辅助函数
func getReaderInfo(readerID int64) *models.ReaderInfo {
	var readerInfo models.ReaderInfo
	if err := utils.DB.First(&readerInfo, readerID).Error; err != nil {
		fmt.Println("Error fetching reader info:", err)
		return nil
	}
	return &readerInfo
}
