package controllers

import (
	"LibrarySystemGolang/models"
	"LibrarySystemGolang/utils"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"strconv"
)

// 获取读者已预约的图书ID列表及状态
func getMyReserveList(readerID int64) []struct {
	BookID   int64
	Accepted bool
} {
	var reserveList []models.Reserve
	var myReserveList []struct {
		BookID   int64
		Accepted bool
	}

	// 查询读者的预约记录
	if err := utils.DB.Where("reader_id = ?", readerID).Find(&reserveList).Error; err != nil {
		fmt.Println("Error fetching reserve list:", err)
		return nil
	}

	// 提取图书ID和预约状态
	for _, record := range reserveList {
		accepted := (!record.AcceptDate.IsZero()) // 如果 accept_date 不为空，则为 true
		myReserveList = append(myReserveList, struct {
			BookID   int64
			Accepted bool
		}{
			BookID:   record.BookID,
			Accepted: accepted,
		})
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
	// 获取分页参数
	page, size := getPageAndSize(c)

	// 调用 queryBook 获取图书数据和总记录数
	books, total := queryBook(c)

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

	// 获取读者的借阅和预约信息
	myRequireReserveMap := make(map[int64]bool)
	myAcceptReserveMap := make(map[int64]bool)
	for _, reserve := range getMyReserveList(readerID) {
		if reserve.Accepted {
			myAcceptReserveMap[reserve.BookID] = true
		} else {
			myRequireReserveMap[reserve.BookID] = true
		}
	}
	// 计算总页数
	totalPages := (total + int64(size) - 1) / int64(size)
	// 准备模板数据
	data := gin.H{
		"books":               books,
		"myLendMap":           myLendMap,
		"myRequireReserveMap": myRequireReserveMap,
		"myAcceptReserveMap":  myAcceptReserveMap,
		"currentPage":         page,
		"totalPages":          totalPages,
		"pageSize":            size,
		"searchField":         c.Query("search_field"),
		"searchKeyword":       c.Query("search_keyword"),
		"prevPage":            page - 1,
		"nextPage":            page + 1,
		"hasPrev":             page > 1,
		"hasNext":             page < int(totalPages),
	}

	// 渲染模板
	if len(books) > 0 {
		c.HTML(http.StatusOK, "reader_book.html", data)
	} else {
		c.HTML(http.StatusOK, "reader_book.html", gin.H{"error": "没有匹配的图书"})
	}
}
func BookHot(c *gin.Context) {

	// 获取分页参数
	page, size := getPageAndSize(c)
	keyWord := c.Query("classID")
	classID := 0
	if keyWord != "" {
		// 获取分类ID
		classID, _ = strconv.Atoi(keyWord)
	}

	// 获取所有分类信息
	var categories []models.ClassInfo
	if err := utils.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取分类信息"})
		return
	}

	// 定义热门图书结构
	type HotBook struct {
		Book      models.Book `json:"book"`
		LendCount int         `json:"lend_count"`
	}
	// 创建一个映射，将 ClassID 映射到 ClassName
	classMap := make(map[int64]string)
	for _, category := range categories {
		classMap[category.ClassID] = category.ClassName
	}

	// 查询借阅数量前5名的图书对象
	var results []struct {
		BookID       int64            `json:"book_id"`
		Name         string           `json:"name"`
		Author       string           `json:"author"`
		Publish      string           `json:"publish"`
		ISBN         string           `json:"isbn"`
		Introduction string           `json:"introduction"`
		Language     string           `json:"language"`
		Price        float64          `json:"price"`
		PubDate      string           `json:"pub_date"`
		ClassID      int64            `json:"class_id"`
		Number       int              `json:"number"`
		Image        string           `json:"image"`
		LendCount    int              `json:"lend_count"`
		ClassInfo    models.ClassInfo `gorm:"foreignKey:ClassID;references:ClassID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	}

	var total int64
	var db = utils.DB.Model(&models.Lend{}).Preload("ClassInfo").
		Select("books.book_id, books.name, books.author, books.publish, books.isbn, books.introduction, books.language, books.price, books.pub_date, books.class_id, books.number, books.image, COUNT(lends.book_id) AS lend_count").
		Joins("JOIN books ON lends.book_id = books.book_id")
	if classID != 0 {
		db = db.Where("books.class_id = ?", classID)
	}
	db = db.Group("books.book_id").Order("COUNT(lends.book_id) DESC")

	// 查询总记录数
	db.Count(&total)
	// 分页查询
	db.Offset((page - 1) * size).Limit(size).Scan(&results)

	// 将查询结果映射到 HotBook 结构体
	var hotBooks []HotBook
	for _, result := range results {
		hotBooks = append(hotBooks, HotBook{
			Book: models.Book{
				BookID:       result.BookID,
				Name:         result.Name,
				Author:       result.Author,
				Publish:      result.Publish,
				ISBN:         result.ISBN,
				Introduction: result.Introduction,
				Language:     result.Language,
				Price:        result.Price,
				PubDate:      result.PubDate,
				ClassID:      result.ClassID,
				Number:       result.Number,
				Image:        result.Image,
				ClassInfo:    result.ClassInfo,
			},
			LendCount: result.LendCount,
		})
	}
	// 计算总页数
	totalPages := (total + int64(size) - 1) / int64(size)
	// 准备模板数据
	data := gin.H{
		"class_info":    categories,
		"lendStatsJSON": queryLends(),
		"class_map":     classMap, // 添加分类映射
		"hot_books":     hotBooks,
		"currentPage":   page,
		"totalPages":    totalPages,
		"pageSize":      size,
		"prevPage":      page - 1,
		"nextPage":      page + 1,
		"hasPrev":       page > 1,
		"hasNext":       page < int(totalPages),
	}
	isAdmin, _ := c.Cookie("isAdmin")
	if isAdmin == "true" {
		c.HTML(http.StatusOK, "admin_book_hot.html", data)
	} else {
		c.HTML(http.StatusOK, "reader_book_hot.html", data)
	}
}

// AdminShowBookPage 获取所有图书
func AdminShowBookPage(c *gin.Context) {
	// 获取分页参数
	page, size := getPageAndSize(c)

	// 调用 queryBook 获取图书数据和总记录数
	books, total := queryBook(c)
	// 计算总页数
	totalPages := (total + int64(size) - 1) / int64(size)

	// 准备模板数据
	data := gin.H{
		"books":         books,
		"currentPage":   page,
		"totalPages":    totalPages,
		"pageSize":      size,
		"searchField":   c.Query("search_field"),
		"searchKeyword": c.Query("search_keyword"),
		"prevPage":      page - 1,
		"nextPage":      page + 1,
		"hasPrev":       page > 1,
		"hasNext":       page < int(totalPages),
	}
	if len(books) > 0 {
		c.HTML(http.StatusOK, "admin_book.html", data)
	} else {
		c.HTML(http.StatusOK, "admin_book.html", gin.H{"error": "没有匹配的图书"})
	}

}
func AdminShowBookHotPage(c *gin.Context) {
	// 获取分页参数
	page, size := getPageAndSize(c)

	// 调用 queryBook 获取图书数据和总记录数
	books, total := queryBook(c)
	// 计算总页数
	totalPages := (total + int64(size) - 1) / int64(size)

	// 准备模板数据
	data := gin.H{
		"books":         books,
		"currentPage":   page,
		"totalPages":    totalPages,
		"pageSize":      size,
		"searchField":   c.Query("search_field"),
		"searchKeyword": c.Query("search_keyword"),
		"prevPage":      page - 1,
		"nextPage":      page + 1,
		"hasPrev":       page > 1,
		"hasNext":       page < int(totalPages),
	}
	if len(books) > 0 {
		c.HTML(http.StatusOK, "admin_book.html", data)
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

		// 检查 ISBN 是否已经存在
		var existingBook models.Book
		if err := utils.DB.Where("isbn = ?", book.ISBN).First(&existingBook).Error; err == nil {
			// 如果存在，更新数量
			existingBook.Number += book.Number
			if err := utils.DB.Save(&existingBook).Error; err != nil {
				log.Printf("更新第 %d 条记录失败: %v", i+1, err)
				continue
			}
		} else {
			// 如果不存在，插入新记录
			if err := utils.DB.Create(&book).Error; err != nil {
				log.Printf("导入第 %d 条记录失败: %v", i+1, err)
				continue
			}
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
func getPageAndSize(c *gin.Context) (int, int) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}
	size, _ := strconv.Atoi(c.Query("size"))
	if size <= 0 {
		size = 20 // 默认每页显示20条
	}
	return page, size
}

func queryBook(c *gin.Context) ([]models.Book, int64) {
	// 获取查询参数
	searchField := c.Query("search_field")
	searchKeyword := c.Query("search_keyword")

	page, size := getPageAndSize(c)

	// 默认查询所有图书
	var books []models.Book
	var total int64

	db := utils.DB.Model(&models.Book{}).Preload("ClassInfo")

	// 如果有查询关键字，则添加查询条件
	if searchKeyword != "" {
		switch searchField {
		case "class_name":
			db = db.Joins("JOIN class_infos ON books.class_id = class_infos.class_id").
				Where("class_infos.class_name LIKE ?", "%"+searchKeyword+"%")
		default:
			querySql := fmt.Sprintf("%s LIKE ?", searchField)
			db = db.Where(querySql, "%"+searchKeyword+"%")
		}
	}

	// 查询总记录数
	db.Count(&total)

	// 分页查询
	db.Offset((page - 1) * size).Limit(size).Find(&books)

	return books, total
}
func queryLends() string {

	// 查询借阅表并统计每种图书分类的借阅量
	var lendStats []struct {
		ClassName string `json:"class_name"`
		Count     int    `json:"count"`
	}
	if err := utils.DB.Table("lends").
		Select("class_infos.class_name AS class_name, COUNT(*) AS count").
		Joins("JOIN books ON lends.book_id = books.book_id").
		Joins("JOIN class_infos ON books.class_id = class_infos.class_id").
		Group("class_infos.class_name").
		Order("count DESC"). // 按借阅量倒序排列
		Scan(&lendStats).Error; err != nil {
		fmt.Println("Error querying lend statistics:", err)
	}

	// 将 lendStats 转换为 JSON 字符串
	lendStatsJSON, _ := json.Marshal(lendStats)
	return string(lendStatsJSON)
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
