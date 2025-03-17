package controllers

import (
	"LibrarySystemGolang/models"
	"LibrarySystemGolang/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

func AdminReserveList(c *gin.Context) {
	var reserve []models.Reserve
	if err := utils.DB.Preload("Book").Preload("ReaderInfo").Find(&reserve).Error; err != nil {
		log.Println("Error fetching reserve list:", err)
		c.HTML(http.StatusOK, "admin_reserve.html", gin.H{"error": "无法获取预约记录"})
		return
	}
	c.HTML(http.StatusOK, "admin_reserve.html", gin.H{"reserve": reserve})
}

// AdminLendList 获取所有借阅记录
func AdminLendList(c *gin.Context) {
	var lends []models.Lend
	if err := utils.DB.Preload("Book").Preload("ReaderInfo").Find(&lends).Error; err != nil {
		log.Println("Error fetching lend list:", err)
		c.HTML(http.StatusOK, "admin_lend.html", gin.H{"error": "无法获取借阅记录"})
		return
	}
	c.HTML(http.StatusOK, "admin_lend.html", gin.H{"lends": lends})
}

// ReaderLend 获取当前读者的借阅记录
func ReaderLend(c *gin.Context) {
	readerIDStr, err := c.Cookie("readercard")
	if err != nil {
		c.HTML(http.StatusOK, "reader_lend.html", gin.H{"error": "未找到读者信息"})
		return
	}
	readerID, _ := strconv.ParseInt(readerIDStr, 10, 64)

	var lends []models.Lend
	if err := utils.DB.Preload("Book").Preload("ReaderInfo").Where("reader_id = ?", readerID).Find(&lends).Error; err != nil {
		log.Println("Error fetching my lend list:", err)
		c.HTML(http.StatusOK, "reader_lend.html", gin.H{"error": "无法获取借阅记录"})
		return
	}
	data := gin.H{
		"lends":         lends,
		"lendStatsJSON": queryReaderLends(readerID),
	}
	c.HTML(http.StatusOK, "reader_lend.html", data)
}
func queryReaderLends(readerID int64) string {
	// 查询个人名下的借阅记录，并按图书分类进行统计
	var lendStats []struct {
		ClassName string `json:"class_name"`
		Count     int    `json:"count"`
	}

	if err := utils.DB.Table("lends").
		Select("class_infos.class_name AS class_name, COUNT(*) AS count").
		Joins("JOIN books ON lends.book_id = books.book_id").
		Joins("JOIN class_infos ON books.class_id = class_infos.class_id").
		Where("lends.reader_id = ?", readerID).
		Group("class_infos.class_name").
		Order("count DESC").
		Scan(&lendStats).Error; err != nil {
		fmt.Println("Error querying lend statistics:", err)
	}

	// 将 lendStats 转换为 JSON 字符串
	lendStatsJSON, _ := json.Marshal(lendStats)
	return string(lendStatsJSON)
}
func AdminReserveAccept(c *gin.Context) {
	serNumStr := c.Param("id")
	serNum, err := strconv.ParseInt(serNumStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的序列号"})
		return
	}

	// 检查读者是否已经借阅了这本书
	var reserve models.Reserve
	if err := utils.DB.Where("ser_num = ?", serNum).First(&reserve).Error; err != nil {
		log.Println("Error fetching reserve record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "未找到预约记录"})
		return
	}
	reserve.SerNum = serNum
	// 更新借阅记录的归还日期
	reserve.AcceptDate = models.LocalDate(time.Now())
	if err := utils.DB.Where(reserve.SerNum).Save(&reserve).Error; err != nil {
		log.Println("Error updating reserve record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "预约通过失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "预约通过"})
}
func AdminReserveDelete(c *gin.Context) {
	serNumStr := c.Param("id")
	serNum, err := strconv.ParseInt(serNumStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的序列号"})
		return
	}

	if err = utils.DB.Where("ser_num = ?", serNum).Delete(&models.Reserve{}).Error; err != nil {
		log.Println("Error deleting reserve record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除预约记录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "预约记录删除成功"})
}

// AdminLendDelete 删除借阅记录
func AdminLendDelete(c *gin.Context) {
	serNumStr := c.Param("id")
	serNum, err := strconv.ParseInt(serNumStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的序列号"})
		return
	}

	if err = utils.DB.Where("ser_num = ?", serNum).Delete(&models.Lend{}).Error; err != nil {
		log.Println("Error deleting lend record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除借阅记录失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "记录删除成功"})
}

// ReaderLendBook 借阅书籍
func ReaderLendBook(c *gin.Context) {
	bookIDStr := c.Param("id")
	readerIDStr, err := c.Cookie("readercard")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到读者信息"})
		return
	}
	readerID, _ := strconv.ParseInt(readerIDStr, 10, 64)
	bookID, _ := strconv.ParseInt(bookIDStr, 10, 64)
	// 检查书籍是否存在
	var book models.Book
	if err := utils.DB.First(&book, bookID).Error; err != nil {
		log.Println("Error fetching book:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "书籍不存在"})
		return
	}

	// 检查读者是否已经借阅了这本书
	var lend models.Lend
	if err := utils.DB.Where("book_id = ? AND reader_id = ? AND back_date IS NULL", bookID, readerID).First(&lend).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "您已经借阅了这本书"})
		return
	}
	// 插入借阅记录
	lend = models.Lend{
		BookID:   bookID,
		ReaderID: readerID,
		LendDate: models.LocalDate(time.Now()),
		BackDate: models.LocalDate{},
	}
	if err := utils.DB.Create(&lend).Error; err != nil {
		log.Println("Error creating lend record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "借阅失败"})
		return
	}

	// 更新书籍数量
	book.Number -= 1
	if err := utils.DB.Where("book_id=?", bookID).Save(&book).Error; err != nil {
		log.Println("Error updating book number:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "借阅失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "借阅成功"})
}

// ReaderReturnBook 归还书籍
func ReaderReturnBook(c *gin.Context) {
	bookIDStr := c.Param("id")
	readerIDStr, err := c.Cookie("readercard")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到读者信息"})
		return
	}
	readerID, _ := strconv.ParseInt(readerIDStr, 10, 64)
	bookID, _ := strconv.ParseInt(bookIDStr, 10, 64)

	// 检查借阅记录是否存在
	var lend models.Lend
	if err := utils.DB.Where("book_id = ? AND reader_id = ? AND back_date IS NULL", bookID, readerID).First(&lend).Error; err != nil {
		log.Println("Error fetching lend record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "未找到借阅记录"})
		return
	}
	// 更新借阅记录的归还日期
	lend.BackDate = models.LocalDate(time.Now())
	if err := utils.DB.Where(lend.SerNum).Save(&lend).Error; err != nil {
		log.Println("Error updating lend record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "归还失败"})
		return
	}

	// 更新书籍数量
	var book models.Book
	if err := utils.DB.First(&book, bookID).Error; err != nil {
		log.Println("Error fetching book:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "归还失败"})
		return
	}
	book.Number += 1
	if err := utils.DB.Where(bookID).Save(&book).Error; err != nil {
		log.Println("Error updating book number:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "归还失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "归还成功"})
}

// 预约书籍
func ReaderReservationBook(c *gin.Context) {
	bookIDStr := c.Param("id")
	readerIDStr, err := c.Cookie("readercard")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "未找到读者信息"})
		return
	}
	readerID, _ := strconv.ParseInt(readerIDStr, 10, 64)
	bookID, _ := strconv.ParseInt(bookIDStr, 10, 64)
	// 检查书籍是否存在
	var book models.Book
	if err := utils.DB.First(&book, bookID).Error; err != nil {
		log.Println("Error fetching book:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "书籍不存在"})
		return
	}
	// 检查读者是否已经借阅了这本书
	var reserve models.Reserve
	if err := utils.DB.Where("book_id = ? AND reader_id = ?", bookID, readerID).First(&reserve).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "您已经预约了这本书"})
		return
	}
	// 插入借阅记录
	reserve = models.Reserve{
		BookID:      bookID,
		ReaderID:    readerID,
		RequireDate: models.LocalDate(time.Now()),
		AcceptDate:  models.LocalDate{},
	}
	if err := utils.DB.Create(&reserve).Error; err != nil {
		log.Println("Error creating reserve record:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "预约失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "预约成功"})
}
