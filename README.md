## Latihan Implementasi

### Demo: `POST /users` (Create User)

Bagian ini mendemonstrasikan alur pembuatan satu endpoint utuh, dari mendaftarkan route sampai data tersimpan ke database.

**`database/entities/user_entity.go`** -- Skema tabel
```go
package entities

type User struct {
    Common                    // embed ID, CreatedAt, UpdatedAt, DeletedAt
    Name     string `gorm:"not null"`
    Email    string `gorm:"unique;not null"`
    Password string `gorm:"not null"`
    Role     string `gorm:"default:'user'"`
}
```

> [!NOTE]
> - `Common` -- Ini disebut **struct embedding**. Struct `User` "mewarisi" semua field dari struct `Common` (biasanya berisi `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`). Mirip konsep inheritance tapi tanpa polymorphism.
> - `gorm:"default:'user'"` -- Jika field `Role` tidak diisi saat membuat user baru, database akan otomatis mengisinya dengan `'user'`.

**`modules/user/dto/user_dto.go`** -- Shape data masuk & keluar
```go
package dto

type CreateUserRequest struct {
    Name     string `json:"name"     binding:"required"`
    Email    string `json:"email"    binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type UserResponse struct {
    ID    uint   `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Role  string `json:"role"`
}
```

> [!NOTE]
> - `binding:"required"` -- Tag dari Gin yang berarti field ini **wajib diisi**. Jika client tidak mengirimnya, Gin akan otomatis mengembalikan error.
> - `binding:"required,email"` -- Selain wajib, nilainya juga **harus berformat email** yang valid.
> - `binding:"required,min=8"` -- Selain wajib, panjang string **minimal 8 karakter**.
> - `CreateUserRequest` digunakan untuk **menerima** data dari client, sedangkan `UserResponse` digunakan untuk **mengirim** data kembali ke client (tanpa password!).

**`modules/user/validation/user_validation.go`** -- Validasi input
```go
package validation

import (
    "github.com/gin-gonic/gin"
    "github.com/username/boilerplate/modules/user/dto"
)

func ValidateCreateUser(c *gin.Context) (*dto.CreateUserRequest, error) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        return nil, err
    }
    return &req, nil
}
```

> [!NOTE]
> - `c.ShouldBindJSON(&req)` -- Membaca body request (JSON), lalu **mengisinya ke struct** `req`. Jika JSON tidak valid atau field yang `required` kosong, fungsi ini mengembalikan error.
> - `var req dto.CreateUserRequest` -- Membuat variabel `req` dengan tipe `CreateUserRequest` (bernilai kosong/default).
> - `return &req, nil` -- Mengembalikan **pointer** ke `req` dan `nil` (tidak ada error). Pointer digunakan agar data tidak perlu di-copy (lebih efisien untuk struct besar).

**`modules/user/repository/user_repository.go`** -- Query database
```go
package repository

import (
    "github.com/username/boilerplate/database/entities"
    "gorm.io/gorm"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *entities.User) error {
    return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*entities.User, error) {
    var user entities.User
    err := r.db.First(&user, id).Error
    return &user, err
}

func (r *UserRepository) FindAll() ([]entities.User, error) {
    var users []entities.User
    err := r.db.Find(&users).Error
    return users, err
}
```

> [!NOTE]
> - `func (r *UserRepository) Create(...)` -- Ini disebut **method**. Bedanya dengan fungsi biasa: method terikat ke sebuah struct. `(r *UserRepository)` disebut **receiver** -- artinya method `Create` milik struct `UserRepository`, dan bisa diakses via `r`.
> - `r.db.Create(user)` -- Method GORM untuk insert data ke database. GORM otomatis membuat query `INSERT INTO users ...` dari struct yang diberikan.
> - `r.db.First(&user, id)` -- Mengambil satu record pertama berdasarkan primary key (`id`). Mirip `SELECT * FROM users WHERE id = ? LIMIT 1`.
> - `r.db.Find(&users)` -- Mengambil semua record. Mirip `SELECT * FROM users`.
> - `.Error` -- Setiap operasi GORM mengembalikan result yang memiliki field `.Error`. Jika operasi berhasil, nilainya `nil`.
> - `[]entities.User` -- Tanda `[]` di depan tipe artinya **slice** (array dinamis). Jadi `[]entities.User` berarti "kumpulan User."

**`modules/user/service/user_service.go`** -- Business logic
```go
package service

import (
    "github.com/username/boilerplate/database/entities"
    "github.com/username/boilerplate/modules/user/dto"
    "github.com/username/boilerplate/modules/user/repository"
    "github.com/username/boilerplate/pkg/helpers"
)

type UserService struct {
    repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
    return &UserService{repo: repo}
}

func (s *UserService) CreateUser(req *dto.CreateUserRequest) (*entities.User, error) {
    // Business logic: hash password sebelum disimpan
    hashedPassword, err := helpers.HashPassword(req.Password)
    if err != nil {
        return nil, err
    }

    user := &entities.User{
        Name:     req.Name,
        Email:    req.Email,
        Password: hashedPassword,
    }

    err = s.repo.Create(user)
    return user, err
}
```

> [!NOTE]
> - `helpers.HashPassword(req.Password)` -- Mengubah password plain text menjadi **hash** (string acak yang tidak bisa di-decode balik). Ini adalah praktik keamanan standar -- password **tidak boleh** disimpan dalam bentuk plain text di database.
> - Service layer **tidak boleh tahu** tentang HTTP atau database secara langsung. Ia hanya menerima data (DTO), memproses business logic, dan memanggil repository untuk operasi database.
> - Pola `New...()` (seperti `NewUserService`, `NewUserRepository`) adalah **constructor pattern** di Go -- fungsi yang membuat dan mengembalikan instance baru dari sebuah struct.

**`modules/user/controller/user_controller.go`** -- Handle HTTP
```go
package controller

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/username/boilerplate/modules/user/service"
    "github.com/username/boilerplate/modules/user/validation"
    "github.com/username/boilerplate/pkg/utils"
)

type UserController struct {
    service *service.UserService
}

func NewUserController(service *service.UserService) *UserController {
    return &UserController{service: service}
}

func (ctrl *UserController) CreateUser(c *gin.Context) {
    req, err := validation.ValidateCreateUser(c)
    if err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
        return
    }

    user, err := ctrl.service.CreateUser(req)
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal membuat user")
        return
    }

    utils.SuccessResponse(c, http.StatusCreated, "User berhasil dibuat", user)
}
```

> [!NOTE]
> - `http.StatusBadRequest` = kode 400 (request dari client salah/tidak valid).
> - `http.StatusInternalServerError` = kode 500 (error di sisi server).
> - `http.StatusCreated` = kode 201 (data berhasil dibuat).
> - `err.Error()` -- Mengonversi objek error menjadi string pesan error yang bisa dibaca.
> - `utils.ErrorResponse` dan `utils.SuccessResponse` -- Fungsi helper untuk memastikan semua response API memiliki **format yang seragam** (konsisten).

**`modules/user/routes.go`** -- Daftarkan endpoint
```go
package user

import (
    "github.com/gin-gonic/gin"
    "github.com/username/boilerplate/middlewares"
    "github.com/username/boilerplate/modules/user/controller"
)

func RegisterUserRoutes(r *gin.RouterGroup, ctrl *controller.UserController) {
    users := r.Group("/users")
    {
        users.POST("", ctrl.CreateUser)                                          // POST /api/users
        users.GET("", middlewares.Authentication(), ctrl.GetAllUsers)            // GET  /api/users (protected)
        users.GET("/:id", middlewares.Authentication(), ctrl.GetUserByID)        // GET  /api/users/:id (protected)
    }
}
```

> [!NOTE]
> - `r.Group("/users")` -- Membuat **route group**. Semua endpoint di dalam group ini otomatis diawali dengan `/users`.
> - `users.POST("", ...)` -- Karena sudah di group `/users`, endpoint ini menjadi `POST /users`.
> - `users.GET("/:id", ...)` -- `:id` disebut **URL parameter**. Nilainya dinamis -- misalnya `/users/5` berarti `id = 5`.
> - `middlewares.Authentication()` -- Middleware yang mengecek JWT token. Endpoint yang dibungkus middleware ini hanya bisa diakses oleh user yang sudah login. `POST /users` (registrasi) tidak memerlukan login, jadi tanpa middleware.

> [!IMPORTANT]
> Perhatikan `middlewares.Authentication()` -- endpoint yang membutuhkan login dibungkus middleware ini. Middleware akan memvalidasi JWT token sebelum request diteruskan ke controller.

---

### Challenge A -- `GET /users/:id`

> Ambil satu user berdasarkan ID. Kembalikan `404` jika tidak ditemukan.

`modules/user/service/user_service.go`
```go
func (s *UserService) GetUserByID(id string) (*entities.User, error) {
	return s.repo.GetByID(id)
}
```

> - `id string` -- Parameter ID yang diterima dari controller.
> - `return s.repo.GetByID(id)` -- Memanggil method repository untuk mengambil data user berdasarkan ID. Hasilnya langsung dikembalikan ke controller.

`modules/user/controller/user_controller.go`
```go
func (ctrl *UserController) GetUserByID(c *gin.Context) {
	id := c.Param("id")
	user, err := ctrl.service.GetUserByID(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Data user berhasil diambil", user)
}
```

> - `id := c.Param("id")` -- Mengambil nilai `id` dari URL parameter. Misalnya jika endpoint dipanggil dengan `/users/5`, maka `id` akan bernilai `"5"`.
> - `utils.ErrorResponse(c, http.StatusNotFound, "User tidak ditemukan")` -- Jika terjadi error (misalnya user dengan ID tersebut tidak ada), kita mengembalikan response dengan status `404 Not Found` dan pesan error yang jelas.
> - `utils.SuccessResponse(c, http.StatusOK, "Data user berhasil diambil", user)` -- Jika berhasil, kita mengembalikan response dengan status `200 OK`, pesan sukses, dan data user yang ditemukan.

`modules/user/repository/user_repository.go`
```go
func (r *UserRepository) GetByID(id string) (*entities.User, error) {
	var user entities.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return &user, err
}
```

> - `var user entities.User` -- Variabel `user` untuk menampung hasil query.
> - `r.db.Where("id = ?", id).First(&user)` -- Query GORM untuk mencari record pertama di tabel `users` yang memiliki `id` sesuai dengan parameter.
> - `return &user, err` -- Mengembalikan pointer ke `user` dan error. Jika user ditemukan, `err` akan bernilai `nil`. Jika tidak ditemukan, `err` akan berisi informasi error yang bisa digunakan untuk menentukan response di controller.

### Challenge B -- `GET /users`

> Ambil semua user. Kembalikan array JSON.

> [!WARNING]
> **Tips debugging:** Error paling umum di Go adalah lupa menangani `if err != nil`. Kalau ada `panic`, cari baris yang tidak menangani error return-nya. Gunakan `fmt.Println(err)` untuk print error ke terminal.

`modules/user/service/user_service.go`
```go
func (s *UserService) GetAllUsers() ([]entities.User, error) {
	return s.repo.GetAll()
}
```

> `return s.repo.GetAll()` -- Memanggil method repository untuk mengambil semua data user. Hasilnya langsung dikembalikan ke controller.

`modules/user/controller/user_controller.go`
```go
func (ctrl *UserController) GetAllUsers(c *gin.Context) {
	users, err := ctrl.service.GetAllUsers()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data users")
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Data users berhasil diambil", users)
}
```

> - `users, err := ctrl.service.GetAllUsers()` -- Memanggil service untuk mendapatkan semua user. Hasilnya adalah slice (array) of `entities.User` dan error.
> - `utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengambil data users")` -- Jika terjadi error saat mengambil data, kita mengembalikan response dengan status dan pesan error yang jelas.
> - `utils.SuccessResponse(c, http.StatusOK, "Data users berhasil diambil", users)` -- Jika berhasil, kita mengembalikan response dengan status `200 OK`, pesan sukses, dan data users yang ditemukan.

`modules/user/repository/user_repository.go`
```go
func (r *UserRepository) GetAll() ([]entities.User, error) {
	var users []entities.User
	err := r.db.Find(&users).Error
	return users, err
}
```

> - `var users []entities.User` -- Membuat slice `users` untuk menampung hasil query.
> - `r.db.Find(&users)` -- Query GORM untuk mengambil semua record dari tabel `users`. Hasilnya akan diisi ke dalam slice `users`.
> - `return users, err` -- Mengembalikan slice `users` dan error. Jika query berhasil, `err` akan bernilai `nil`.

---
