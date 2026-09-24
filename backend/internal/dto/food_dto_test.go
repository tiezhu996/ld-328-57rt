package dto

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func bindBody(t *testing.T, body string) (*CreateFoodRequest, error) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	var r CreateFoodRequest
	err := c.ShouldBindJSON(&r)
	return &r, err
}

func TestUnitPriceBinding(t *testing.T) {
	base := `"family_id":1,"name":"x","category":"dairy"`
	// 正常单价
	r, err := bindBody(t, `{`+base+`,"unit_price":5.55}`)
	if err != nil || r.UnitPrice == nil || *r.UnitPrice != 5.55 {
		t.Fatalf("valid price: r=%+v err=%v", r.UnitPrice, err)
	}
	// 负数应被 binding 拒绝
	if _, err := bindBody(t, `{`+base+`,"unit_price":-1}`); err == nil {
		t.Fatal("expected binding error for negative price")
	}
	// null → nil（清空/未填）
	r, err = bindBody(t, `{`+base+`,"unit_price":null}`)
	if err != nil || r.UnitPrice != nil {
		t.Fatalf("null price: r=%+v err=%v", r.UnitPrice, err)
	}
	// 缺省 → nil
	r, err = bindBody(t, `{`+base+`}`)
	if err != nil || r.UnitPrice != nil {
		t.Fatalf("absent price: r=%+v err=%v", r.UnitPrice, err)
	}
	// 0 是合法单价
	r, err = bindBody(t, `{`+base+`,"unit_price":0}`)
	if err != nil || r.UnitPrice == nil || *r.UnitPrice != 0 {
		t.Fatalf("zero price: r=%+v err=%v", r.UnitPrice, err)
	}
}
