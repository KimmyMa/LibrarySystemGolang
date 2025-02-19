package controllers

import (
	"LibrarySystemGolang/models"
	"LibrarySystemGolang/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"time"
)

// AdminLendList 获取所有借阅记录
func AdminLendList(c *gin.Context) {
	var lends []models.Lend
	if err := utils.DB.Preload("Book").Preload("ReaderInfo").Find(&lends).Error; err != nil {
		log.Println("Error fetching lend list:", err)
		c.HTML(http.StatusOK, "admin_lend_list.html", gin.H{"error": "无法获取借阅记录"})
		return
	}
	c.HTML(http.StatusOK, "admin_lend_list.html", gin.H{"lends": lends})
}

// ReaderLend 获取当前读者的借阅记录
func ReaderLend(c *gin.Context) {
	readerIDStr, err := c.Cookie("readercard")
	if err != nil {
		c.HTML(http.StatusOK, "reader_lend_list.html", gin.H{"error": "未找到读者信息"})
		return
	}
	readerID, _ := strconv.ParseInt(readerIDStr, 10, 64)

	var lends []models.Lend
	if err := utils.DB.Preload("Book").Preload("ReaderInfo").Where("reader_id = ?", readerID).Find(&lends).Error; err != nil {
		log.Println("Error fetching my lend list:", err)
		c.HTML(http.StatusOK, "reader_lend_list.html", gin.H{"error": "无法获取借阅记录"})
		return
	}
	c.HTML(http.StatusOK, "reader_lend_list.html", gin.H{"lends": lends})
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
	fmt.Println(book)
	if err := utils.DB.Where(bookID).Save(&book).Error; err != nil {
		log.Println("Error updating book number:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "归还失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "归还成功"})
}
