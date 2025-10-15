package helper

import (
	"personal_schedule_service/internal/collection"
	labels_constant "personal_schedule_service/internal/constant/labels"
	"personal_schedule_service/internal/grpc/utils"
	"time"
)

type LabelHelper struct{}

func (h *LabelHelper) GenerateLabel() []collection.Label {
	now := time.Now()

	labels := []collection.Label{
		// Work Type
		{
			Name:           "Lặp lại",
			Key:            labels_constant.LabelRepeated,
			Meaning:        utils.ToStringPointer("Công việc được thực hiện định kỳ, lặp lại theo chu kỳ"),
			Note:           utils.ToStringPointer("Thường dùng cho các công việc, hay những thói quen hằng ngày"),
			Color:          utils.ToStringPointer("#00C8FF"),
			LabelType:      labels_constant.LabelTypeWorkType,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Hàng Ngày",
			Key:            labels_constant.LabelDaily,
			Meaning:        utils.ToStringPointer("Công việc thực hiện chỉ trong ngày hôm đó"),
			Note:           utils.ToStringPointer("Thường dùng cho các tác vụ giải quyết được trong ngày, không cần kéo dài"),
			Color:          utils.ToStringPointer("#E8E8E8"),
			LabelType:      labels_constant.LabelTypeWorkType,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Nhóm",
			Key:            labels_constant.LabelGroup,
			Meaning:        utils.ToStringPointer("Công việc được đồng bộ từ các nhóm của bạn, sẽ biến mất khi hoàn thành hoặc xóa trong nhóm"),
			Note:           utils.ToStringPointer("Công việc nhóm của bạn, được xử lý ở chức năng quản lý lịch trình nhóm"),
			Color:          utils.ToStringPointer("#FF5EEF"),
			LabelType:      labels_constant.LabelTypeWorkType,
			CreatedAt:      now,
			LastModifiedAt: now,
		},

		// Status
		{
			Name:           "Chờ Làm",
			Key:            labels_constant.LabelPending,
			Meaning:        utils.ToStringPointer("Công việc đã được lên kế hoạch nhưng chưa bắt đầu"),
			Note:           utils.ToStringPointer("Đang chờ điều kiện phù hợp hoặc tài nguyên để bắt đầu"),
			Color:          utils.ToStringPointer("#FFEA00"),
			LabelType:      labels_constant.LabelTypeStatus,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Đang làm",
			Key:            labels_constant.LabelInProgress,
			Meaning:        utils.ToStringPointer("Công việc đang được thực hiện"),
			Note:           utils.ToStringPointer("Đang trong quá trình xử lý, cần theo dõi tiến độ"),
			Color:          utils.ToStringPointer("#00C8FF"),
			LabelType:      labels_constant.LabelTypeStatus,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Hoàn Thành",
			Key:            labels_constant.LabelCompleted,
			Meaning:        utils.ToStringPointer("Công việc đã được hoàn thành thành công"),
			Note:           utils.ToStringPointer("Đã đạt được mục tiêu và yêu cầu đề ra"),
			Color:          utils.ToStringPointer("#00FF00"),
			LabelType:      labels_constant.LabelTypeStatus,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Quá Hạn",
			Key:            labels_constant.LabelOverDue,
			Meaning:        utils.ToStringPointer("Công việc đã vượt quá thời hạn quy định (nhưng vẫn nằm trong ngày)"),
			Note:           utils.ToStringPointer("Là các công việc quá hạn thời gian nhưng chưa đánh dấu hoàn thành, có thể đánh dấu hoàn thành sau nếu vẫn còn trong ngày"),
			Color:          utils.ToStringPointer("#FF7C7E"),
			LabelType:      labels_constant.LabelTypeStatus,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Bỏ cuộc",
			Key:            labels_constant.LabelGiveUp,
			Meaning:        utils.ToStringPointer("Công việc đã hết ngày nhưng chưa hoàn thành"),
			Note:           utils.ToStringPointer("Hệ thống tự động đánh dấu các công việc quá hạn và hết ngày"),
			Color:          utils.ToStringPointer("#ED1E02"),
			LabelType:      labels_constant.LabelTypeStatus,
			CreatedAt:      now,
			LastModifiedAt: now,
		},

		// Difficulty
		{
			Name:           "Dễ",
			Key:            labels_constant.LabelDifficultyEasy,
			Meaning:        utils.ToStringPointer("Công việc đơn giản, không yêu cầu kỹ năng cao và công sức nhiều"),
			Note:           utils.ToStringPointer("Có thể hoàn thành nhanh chóng, dưới 1h"),
			Color:          utils.ToStringPointer("#13C540"),
			LabelType:      labels_constant.LabelTypeDifficulty,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Trung Bình",
			Key:            labels_constant.LabelDifficultyMedium,
			Meaning:        utils.ToStringPointer("Công việc có độ phức tạp vừa phải"),
			Note:           utils.ToStringPointer("Cần có kinh nghiệm và kỹ năng nhất định để thực hiện, Thường tốn vừa phải thời gian 1 - 3h để hoàn thành"),
			Color:          utils.ToStringPointer("#FFEA00"),
			LabelType:      labels_constant.LabelTypeDifficulty,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Khó",
			Key:            labels_constant.LabelDifficultyHard,
			Meaning:        utils.ToStringPointer("Công việc yêu cầu sự tập trung cao hoặc thời gian lớn"),
			Note:           utils.ToStringPointer("Cần nhiều thời gian, tài nguyên và chuyên môn sâu, thời gian hoàn thành thường là trên 3h"),
			Color:          utils.ToStringPointer("#FF5A43"),
			LabelType:      labels_constant.LabelTypeDifficulty,
			CreatedAt:      now,
			LastModifiedAt: now,
		},

		// Priority
		{
			Name:           "Quan trọng & Khẩn cấp",
			Key:            labels_constant.LabelPriorityImportantUrgent,
			Meaning:        utils.ToStringPointer("Ưu tiên cao nhất - cần xử lý ngay lập tức"),
			Note:           utils.ToStringPointer("Thường là các vấn đề khẩn cấp, sự cố cần giải quyết ngay"),
			Color:          utils.ToStringPointer("#FF5A43"),
			LabelType:      labels_constant.LabelTypePriority,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Quan trọng & Không khẩn cấp",
			Key:            labels_constant.LabelPriorityImportantNotUrgent,
			Meaning:        utils.ToStringPointer("Quan trọng nhưng có thể lên kế hoạch thực hiện"),
			Note:           utils.ToStringPointer("Các mục tiêu dài hạn, phát triển chiến lược"),
			Color:          utils.ToStringPointer("#13C540"),
			LabelType:      labels_constant.LabelTypePriority,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Không quan trọng & Khẩn cấp",
			Key:            labels_constant.LabelPriorityNotImportantUrgent,
			Meaning:        utils.ToStringPointer("Cần xử lý nhanh nhưng không ảnh hưởng lớn"),
			Note:           utils.ToStringPointer("Có thể ủy quyền hoặc xử lý nhanh gọn"),
			Color:          utils.ToStringPointer("#FFEA00"),
			LabelType:      labels_constant.LabelTypePriority,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Không quan trọng & Không Khẩn cấp",
			Key:            labels_constant.LabelPriorityNotImportantNotUrgent,
			Meaning:        utils.ToStringPointer("Ưu tiên thấp nhất - có thể hoãn lại"),
			Note:           utils.ToStringPointer("Nên hạn chế thời gian dành cho loại công việc này"),
			Color:          utils.ToStringPointer("#B9B9B9"),
			LabelType:      labels_constant.LabelTypePriority,
			CreatedAt:      now,
			LastModifiedAt: now,
		},

		// Category
		{
			Name:           "Công việc",
			Key:            labels_constant.LabelCategoryWork,
			Meaning:        utils.ToStringPointer("Nhiệm vụ liên quan đến nghề nghiệp, công ty, kinh doanh hoặc dự án đang làm"),
			Note:           utils.ToStringPointer("Bao gồm họp, dự án, báo cáo, và các hạn chót."),
			Color:          utils.ToStringPointer("#3B82F6"),
			LabelType:      labels_constant.LabelTypeCategory,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Cá nhân",
			Key:            labels_constant.LabelCategoryPersonal,
			Meaning:        utils.ToStringPointer("Hoạt động riêng tư ngoài công việc"),
			Note:           utils.ToStringPointer("Sở thích, mục tiêu cá nhân, chăm sóc bản thân"),
			Color:          utils.ToStringPointer("#22C55E"),
			LabelType:      labels_constant.LabelTypeCategory,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Học tập",
			Key:            labels_constant.LabelCategoryStudy,
			Meaning:        utils.ToStringPointer("Các hoạt động học hỏi, nghiên cứu, hoặc đào tạo."),
			Note:           utils.ToStringPointer("Dành cho học sinh, sinh viên, khóa học chuyên môn, hoặc tự học."),
			Color:          utils.ToStringPointer("#A855F7"),
			LabelType:      labels_constant.LabelTypeCategory,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Gia đình",
			Key:            labels_constant.LabelCategoryFamily,
			Meaning:        utils.ToStringPointer("Công việc nhà và trách nhiệm trong gia đình"),
			Note:           utils.ToStringPointer("Dọn dẹp, sửa chữa, mua sắm, nấu ăn, việc vặt hằng ngày"),
			Color:          utils.ToStringPointer("#FB923C"),
			LabelType:      labels_constant.LabelTypeCategory,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Tài chính",
			Key:            labels_constant.LabelCategoryFinance,
			Meaning:        utils.ToStringPointer("Quản lý tiền bạc và công việc hành chính"),
			Note:           utils.ToStringPointer("Thanh toán hóa đơn, lập ngân sách, thuế/ giấy tờ"),
			Color:          utils.ToStringPointer("#FACC15"),
			LabelType:      labels_constant.LabelTypeCategory,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Sức khỏe",
			Key:            labels_constant.LabelCategoryHealth,
			Meaning:        utils.ToStringPointer("Các hoạt động chăm sóc thể chất và tinh thần"),
			Note:           utils.ToStringPointer("Tập thể dục, khám bệnh, thiền, hoặc chăm sóc sức khỏe"),
			Color:          utils.ToStringPointer("#F43F5E"),
			LabelType:      labels_constant.LabelTypeCategory,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Xã hội",
			Key:            labels_constant.LabelCategorySocial,
			Meaning:        utils.ToStringPointer("Các hoạt động chăm sóc thể chất và tinh thần"),
			Note:           utils.ToStringPointer("Tập thể dục, khám bệnh, thiền, hoặc chăm sóc sức khỏe"),
			Color:          utils.ToStringPointer("#06B6D4"),
			LabelType:      labels_constant.LabelTypeCategory,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
		{
			Name:           "Di chuyển",
			Key:            labels_constant.LabelCategoryTravel,
			Meaning:        utils.ToStringPointer("Việc đi lại từ nơi này đến nơi khác"),
			Note:           utils.ToStringPointer("Đi làm, công tác, du lịch, hoặc chạy việc vặt"),
			Color:          utils.ToStringPointer("#9CA3AF"),
			LabelType:      labels_constant.LabelTypeCategory,
			CreatedAt:      now,
			LastModifiedAt: now,
		},

		// Draft
		{
			Name:           "Bản Nháp",
			Key:            labels_constant.LabelDraft,
			Meaning:        utils.ToStringPointer("Công việc chưa được lưu trữ"),
			Note:           utils.ToStringPointer("Công việc được sinh ra từ AI hay khôi phục từ hôm trước, nhưng chưa được lưu"),
			Color:          utils.ToStringPointer("#E879F9"),
			LabelType:      labels_constant.LabelTypeDraft,
			CreatedAt:      now,
			LastModifiedAt: now,
		},
	}

	return labels
}
