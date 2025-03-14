package controllers

import (
	"LibrarySystemGolang/models"
	"LibrarySystemGolang/utils"
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strconv"
)

// 获取读者已预约的图书ID列表
func getMyReserveList(readerID int64) []int64 {
	var reserveList []models.Reserve
	var myReserveList []int64

	// 查询读者的借阅记录
	if err := utils.DB.Where("reader_id = ?", readerID).Find(&reserveList).Error; err != nil {
		fmt.Println("Error fetching reserve list:", err)
		return nil
	}

	// 提取图书ID
	for _, record := range reserveList {
		myReserveList = append(myReserveList, record.BookID)
	}

	return myReserveList
}

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
	books := queryBook(c)

	// 从Cookie中获取读者ID
	readerIDStr, err := c.Cookie("readercard")
	if err != nil {
		c.HTML(http.StatusOK, "reader_book.html", gin.H{"error": "未找到读者信息"})
		return
	}
	readerID, _ := strconv.ParseInt(readerIDStr, 10, 64)

	myLendMap := make(map[int64]bool)
	for _, bookID := range getMyLendList(readerID) {
		myLendMap[bookID] = true
	}

	myReserveMap := make(map[int64]bool)
	for _, bookID := range getMyReserveList(readerID) {
		myReserveMap[bookID] = true
	}
	if len(books) > 0 {
		c.HTML(http.StatusOK, "reader_book.html", gin.H{
			"books":        books,
			"myLendMap":    myLendMap, // 确保传递到模板中
			"myReserveMap": myReserveMap,
		})
	} else {
		c.HTML(http.StatusOK, "reader_book.html", gin.H{"error": "没有匹配的图书"})
	}
}

// AdminShowBookPage 获取所有图书
func AdminShowBookPage(c *gin.Context) {
	books := queryBook(c)
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
func AdminShowBookImportPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_book_import.html", gin.H{"success": ""})
}

func AdminBookImport(c *gin.Context) {
	file, err := c.FormFile("bookFile") // 获取导入的文件
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件导入失败"})
		return
	}

	// 保存文件到服务器
	filePath := fmt.Sprintf("./uploads/%s", file.Filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件保存失败"})
		return
	}

	// 打开 CSV 文件
	csvFile, err := os.Open(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法打开文件"})
		return
	}
	defer csvFile.Close()

	// 读取 CSV 文件
	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1 // 允许字段数量不一致
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取 CSV 文件失败"})
		return
	}

	// 解析 CSV 数据并导入数据库
	for i, record := range records {
		// 跳过表头
		if i == 0 {
			continue
		}

		// 解析 CSV 记录
		book := models.Book{
			Name:         record[1],
			Author:       record[2],
			Publish:      record[3],
			ISBN:         record[4],
			Introduction: record[5],
			Language:     record[6],
			Price:        parseFloat(record[7]), // 解析价格
			PubDate:      record[8],
			ClassID:      parseInt64(record[9]), // 解析分类 ID
			Number:       parseInt(record[10]),  // 解析数量
			Image:        record[11],
		}

		// 将图书信息保存到数据库
		if err := utils.DB.Create(&book).Error; err != nil {
			log.Printf("导入第 %d 条记录失败: %v", i+1, err)
			continue
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": "图书信息导入成功"})
}

// 辅助函数：解析字符串为 float64
func parseFloat(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return value
}

// 辅助函数：解析字符串为 int64
func parseInt64(s string) int64 {
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return value
}

// 辅助函数：解析字符串为 int
func parseInt(s string) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return value
}

// AdminBookExport 导出图书信息为 CSV
func AdminBookExport(c *gin.Context) {
	// 从数据库获取所有图书记录
	books := getAllBooks()

	// 设置响应头，告诉浏览器这是一个 CSV 文件
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=导出的图书.csv")

	// 创建 CSV 写入器
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// 写入 CSV 表头
	header := []string{
		"BookID", "Name", "Author", "Publish", "ISBN", "Introduction",
		"Language", "Price", "PubDate", "ClassID", "Number", "Image",
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "写入 CSV 表头失败"})
		return
	}

	// 写入图书数据
	for _, book := range books {
		record := []string{
			strconv.FormatInt(book.BookID, 10), // BookID
			book.Name,                          // Name
			book.Author,                        // Author
			book.Publish,                       // Publish
			book.ISBN,                          // ISBN
			book.Introduction,                  // Introduction
			book.Language,                      // Language
			strconv.FormatFloat(book.Price, 'f', 2, 64), // Price
			book.PubDate,                        // PubDate
			strconv.FormatInt(book.ClassID, 10), // ClassID
			strconv.Itoa(book.Number),           // Number
			book.Image,                          // Image
		}
		if err := writer.Write(record); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "写入 CSV 数据失败"})
			return
		}
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
func queryBook(c *gin.Context) []models.Book {
	// 获取前端传入的查询参数
	searchField := c.Query("search_field")
	searchKeyword := c.Query("search_keyword")
	if searchKeyword == "" {
		return getAllBooks()
	}
	var books []models.Book
	var querySql = fmt.Sprintf("%s LIKE ?", searchField)
	if err := utils.DB.Preload("ClassInfo").Where(querySql, "%"+searchKeyword+"%").Find(&books).Error; err != nil {
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
