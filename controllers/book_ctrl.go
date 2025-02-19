package controllers

import (
	"LibrarySystemGolang/models"
	"LibrarySystemGolang/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

// 获取读者已借阅的图书ID列表
func getMyLendList(readerID int64) []int64 {
	var lendList []models.Lend
	var myLendList []int64

	// 查询读者的借阅记录
	if err := utils.DB.Where("reader_id = ? AND back_date IS NULL", readerID).Find(&lendList).Error; err != nil {
		fmt.Println("Error fetching lend list:", err)
		return nil
	}

	// 提取图书ID
	for _, record := range lendList {
		myLendList = append(myLendList, record.BookID)
	}

	return myLendList
}
func ReaderBook(c *gin.Context) {
	searchWord := c.Query("searchWord")
	books := queryBook(searchWord)

	// 从Cookie中获取读者ID
	readerIDStr, err := c.Cookie("readercard")
	if err != nil {
		c.HTML(http.StatusOK, "reader_book.html", gin.H{"error": "未找到读者信息"})
		return
	}
	readerID, _ := strconv.ParseInt(readerIDStr, 10, 64)
	myLendList := getMyLendList(readerID) // 获取读者已借阅的图书ID列表
	// 将 myLendList 转换为 map，便于前端快速判断
	myLendMap := make(map[int64]bool)
	for _, bookID := range myLendList {
		myLendMap[bookID] = true
	}

	if len(books) > 0 {
		c.HTML(http.StatusOK, "reader_book.html", gin.H{
			"books":     books,
			"myLendMap": myLendMap, // 确保传递到模板中
		})
	} else {
		c.HTML(http.StatusOK, "reader_book.html", gin.H{"error": "没有匹配的图书"})
	}
}

// AdminShowBookPage 获取所有图书
func AdminShowBookPage(c *gin.Context) {

	searchWord := c.Query("searchWord")
	books := queryBook(searchWord)
	if len(books) > 0 {
		c.HTML(http.StatusOK, "admin_book.html", gin.H{"books": books})
	} else {
		c.HTML(http.StatusOK, "admin_book.html", gin.H{"error": "没有匹配的图书"})
	}

}

// AdminShowBookAddPage 显示添加图书页面
func AdminShowBookAddPage(c *gin.Context) {
	var classInfo []models.ClassInfo
	if err := utils.DB.Find(&classInfo).Error; err != nil {
		log.Println("Error fetching class infos:", err)
		c.HTML(http.StatusOK, "admin_book_add.html", gin.H{"error": "无法获取分类记录"})
		return
	}
	c.HTML(http.StatusOK, "admin_book_add.html", gin.H{"class_infos": classInfo})
}

// AdminBookCreate 处理添加图书操作
func AdminBookCreate(c *gin.Context) {
	var book models.Book
	if err := c.ShouldBind(&book); err != nil {
		fmt.Println("绑定数据失败:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "绑定数据失败"})
		return
	}
	if addBook(&book) {
		c.JSON(http.StatusOK, gin.H{"success": "图书添加成功"})
	} else {
		fmt.Println("图书添加失败")
		c.JSON(http.StatusBadRequest, gin.H{"error": "图书添加失败"})
	}
}

// AdminShowBookUpdatePage 显示编辑图书页面
func AdminShowBookUpdatePage(c *gin.Context) {
	bookIdStr := c.Param("id")
	bookId, _ := strconv.ParseInt(bookIdStr, 10, 64)
	book := getBook(bookId)

	var classInfo []models.ClassInfo
	if err := utils.DB.Find(&classInfo).Error; err != nil {
		log.Println("Error fetching class infos:", err)
		c.HTML(http.StatusOK, "admin_book_add.html", gin.H{"error": "无法获取分类记录"})
		return
	}
	if book != nil {
		c.HTML(http.StatusOK, "admin_book_update.html", gin.H{"detail": book, "class_infos": classInfo})
	} else {
		c.HTML(http.StatusBadRequest, "admin_book.html", gin.H{"error": "图书未找到"})
	}
}

// AdminBookUpdate 处理编辑图书操作
func AdminBookUpdate(c *gin.Context) {
	bookIdStr := c.Param("id")
	bookId, _ := strconv.ParseInt(bookIdStr, 10, 64)
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		fmt.Println("Error binding data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "绑定数据失败", "detail": book})
		return
	}

	// 设置图书 ID，确保更新的是正确的记录
	book.BookID = bookId

	// 更新图书信息
	if err := utils.DB.Model(&models.Book{}).Where("book_id = ?", bookId).Updates(&book).Error; err != nil {
		fmt.Println("Error editing book:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "图书修改失败", "detail": book})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "图书修改成功", "detail": book})
}

// AdminBookDetail 显示图书详细信息页面
func AdminBookDetail(c *gin.Context) {
	bookIdStr := c.Param("id")
	bookId, _ := strconv.ParseInt(bookIdStr, 10, 64)
	book := getBook(bookId)
	if book != nil {
		c.HTML(http.StatusOK, "admin_book_detail.html", gin.H{"detail": book})
	} else {
		c.HTML(http.StatusOK, "admin_book.html", gin.H{"error": "图书未找到"})
	}
}

// ReaderBookDetail 显示读者图书详细信息页面
func ReaderBookDetail(c *gin.Context) {
	bookIdStr := c.Param("id")
	bookId, _ := strconv.ParseInt(bookIdStr, 10, 64)
	book := getBook(bookId)
	if book != nil {
		c.HTML(http.StatusOK, "reader_book_detail.html", gin.H{"detail": book})
	} else {
		c.HTML(http.StatusOK, "reader_book.html", gin.H{"error": "图书未找到"})
	}
}

// 辅助函数
func queryBook(name string) []models.Book {
	if name == "" {
		return getAllBooks()
	}
	var books []models.Book
	if err := utils.DB.Preload("ClassInfo").Where("name LIKE ?", "%"+name+"%").Find(&books).Error; err != nil {
		fmt.Println("Error querying books:", err)
	}
	return books
}

func getAllBooks() []models.Book {
	var books []models.Book
	if err := utils.DB.Preload("ClassInfo").Find(&books).Error; err != nil {
		fmt.Println("Error fetching all books:", err)
		return nil
	}
	return books
}

func addBook(book *models.Book) bool {
	if err := utils.DB.Create(book).Error; err != nil {
		fmt.Println("Error adding book:", err)
		return false
	}
	return true
}

func getBook(bookId int64) *models.Book {
	var book models.Book
	if err := utils.DB.Preload("ClassInfo").First(&book, bookId).Error; err != nil {
		fmt.Println("Error fetching book:", err)
		return nil
	}
	return &book
}

func AdminBookDelete(c *gin.Context) {
	bookIdStr := c.Param("id")
	var book models.Book
	bookId, _ := strconv.ParseInt(bookIdStr, 10, 64)
	if err := utils.DB.Preload("ClassInfo").Delete(&book, bookId).Error; err != nil {
		fmt.Println("Error fetching book:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "图书删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "图书删除成功"})
}
