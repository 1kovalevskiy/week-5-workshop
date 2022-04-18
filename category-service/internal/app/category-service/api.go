package category_service

import (
	"github.com/ozon-edu-go-2021/week-5-workshop/category-service/internal/service/category"
	"github.com/ozon-edu-go-2021/week-5-workshop/category-service/internal/service/task"
	desc "github.com/ozon-edu-go-2021/week-5-workshop/category-service/pkg/category-service"
)

type Implementation struct {
	desc.UnimplementedCategoryServiceServer

	categoryService category.Service
	taskService     task.Service
}

func NewCategoryService(
	categoryService category.Service,
	taskService task.Service,
) desc.CategoryServiceServer {
	return &Implementation{
		categoryService: categoryService,
		taskService:     taskService,
	}
}
