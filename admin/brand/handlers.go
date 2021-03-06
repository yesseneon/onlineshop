package brand

import (
	"net/http"
	"strconv"

	"github.com/yesseneon/onlineshop/helper"
)

func Handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := helper.GetContextData(r.Context())
		action := helper.DefineAction(r)
		switch action {
		case "index":
			index(w, r, ctx)
		case "create":
			create(w, r, ctx)
		case "store":
			store(w, r, ctx)
		case "edit":
			edit(w, r, ctx)
		case "update":
			update(w, r, ctx)
		case "destroy":
			destroy(w, r)
		case "notFound":
			http.Error(w, http.StatusText(404), http.StatusNotFound)
		default:
			http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		}
	})
}

func index(w http.ResponseWriter, r *http.Request, ctx helper.ContextData) {
	type Data struct {
		Context helper.ContextData
		Brands  []Brand
	}

	brands, err := AllBrands()
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	data := Data{
		Context: ctx,
		Brands:  brands,
	}
	helper.Render(w, "brand.gohtml", data)
	return
}

func create(w http.ResponseWriter, r *http.Request, ctx helper.ContextData) {
	type Data struct {
		Context helper.ContextData
		Brand   Brand
	}

	data := Data{
		Context: ctx,
	}
	helper.Render(w, "brand_form.gohtml", data)
	return
}

func store(w http.ResponseWriter, r *http.Request, ctx helper.ContextData) {
	brand := &Brand{
		Name: r.FormValue("name"),
	}

	if brand.validate() == false {
		type Data struct {
			Context helper.ContextData
			Brand   *Brand
		}

		data := Data{
			Context: ctx,
			Brand:   brand,
		}
		helper.Render(w, "brand_form.gohtml", data)
		return
	}

	_, err := brand.store()
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/brands/", http.StatusSeeOther)
	return
}

func edit(w http.ResponseWriter, r *http.Request, ctx helper.ContextData) {
	type Data struct {
		Context helper.ContextData
		Brand   Brand
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	brand, err := oneBrand(id)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}
	if brand.ID < 1 {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		return
	}

	data := Data{
		Context: ctx,
		Brand:   brand,
	}
	helper.Render(w, "brand_form.gohtml", data)
	return
}

func update(w http.ResponseWriter, r *http.Request, ctx helper.ContextData) {
	id, err := strconv.Atoi(r.FormValue("_id"))
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	brand := &Brand{
		ID:   id,
		Name: r.FormValue("name"),
	}

	if brand.validate() == false {
		type Data struct {
			Context helper.ContextData
			Brand   *Brand
		}

		data := Data{
			Context: ctx,
			Brand:   brand,
		}
		helper.Render(w, "brand_form.gohtml", data)
		return
	}

	err = brand.update()
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/brands/", http.StatusSeeOther)
	return
}

func destroy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.FormValue("_id"))
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	brand := &Brand{ID: id}
	err = brand.destroy()
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/brands/", http.StatusSeeOther)
	return
}
