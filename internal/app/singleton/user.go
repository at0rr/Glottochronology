package singleton

const currentUserID uint = 1

// CurrentUserID возвращает ID текущего пользователя.
// Используется во всех методах, где нужен создатель.
func CurrentUserID() uint {
	return currentUserID
}
