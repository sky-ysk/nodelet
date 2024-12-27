package rest

import (
	"hit.edu/framework/pkg/apis/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
)

// WipeObjectMetaSystemFields 清除ObjectMeta中系统自动管理的字段
func WipeObjectMetaSystemFields(metas meta.Object) {
	metas.SetCreationTimestamp(meta.Time{})
	metas.SetUID("")
	metas.SetDeletionTimestamp(nil)
	metas.SetDeletionGracePeriodSeconds(nil)
	//metas.SetSelfLink("")
}
func WipeObjectMetaSystemFields1(metas meta.Object) {
	metas.SetCreationTimestamp(meta.Time{})
	metas.SetUID("")
	metas.SetDeletionTimestamp(nil)
	metas.SetDeletionGracePeriodSeconds(nil)
}

// FillObjectMetaSystemFields 填充ObjectMeta中系统自动管理的字段
func FillObjectMetaSystemFields(metas metav1.Object) {
	metas.SetCreationTimestamp(metav1.Now())
	metas.SetUID(uuid.NewUUID())
}
func FillObjectMetaSystemFields1(metas meta.Object) {
	metas.SetCreationTimestamp(meta.Now())
	uid1 := uuid.NewUUID()
	uid := meta.UID(string(uid1)) // 转换为 meta.UID 类型的值
	metas.SetUID(uid)
}
