package models

import (
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTagTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&Tag{}, &Task{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestTagStruct(t *testing.T) {
	tag := Tag{
		Name:  "work",
		Color: "#3b82f6",
	}

	if tag.Name != "work" {
		t.Errorf("Expected name 'work', got %s", tag.Name)
	}

	if tag.Color != "#3b82f6" {
		t.Errorf("Expected color '#3b82f6', got %s", tag.Color)
	}
}

func TestTagBeforeCreate(t *testing.T) {
	db := setupTagTestDB(t)

	t.Run("generates UUID when ID is nil", func(t *testing.T) {
		tag := Tag{Name: "personal", Color: "#fecaca"}

		if tag.ID != uuid.Nil {
			t.Errorf("Expected ID to be nil initially")
		}

		err := tag.BeforeCreate(db)
		if err != nil {
			t.Errorf("BeforeCreate returned error: %v", err)
		}

		if tag.ID == uuid.Nil {
			t.Errorf("Expected ID to be generated")
		}
	})

	t.Run("preserves existing UUID", func(t *testing.T) {
		existingID := uuid.New()
		tag := Tag{
			ID:    existingID,
			Name:  "urgent",
			Color: "#fca5a5",
		}

		err := tag.BeforeCreate(db)
		if err != nil {
			t.Errorf("BeforeCreate returned error: %v", err)
		}

		if tag.ID != existingID {
			t.Errorf("Expected ID to remain %s, got %s", existingID, tag.ID)
		}
	})
}

func TestTagCreate(t *testing.T) {
	db := setupTagTestDB(t)

	t.Run("creates tag with name and color", func(t *testing.T) {
		tag := Tag{Name: "work", Color: "#3b82f6"}

		err := tag.Create(db)
		if err != nil {
			t.Errorf("Create returned error: %v", err)
		}

		if tag.ID == uuid.Nil {
			t.Error("Expected ID to be generated")
		}

		var dbTag Tag
		result := db.First(&dbTag, "id = ?", tag.ID)
		if result.Error != nil {
			t.Errorf("Tag not found in database: %v", result.Error)
		}

		if dbTag.Name != "work" {
			t.Errorf("Expected name 'work', got %s", dbTag.Name)
		}
	})

	t.Run("creates tag without color (assigns random)", func(t *testing.T) {
		tag := Tag{Name: "personal"}

		err := tag.Create(db)
		if err != nil {
			t.Errorf("Create returned error: %v", err)
		}

		if tag.Color == "" {
			t.Error("Expected color to be assigned")
		}

		found := false
		for _, color := range tagColors {
			if tag.Color == color {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected color to be from predefined list, got %s", tag.Color)
		}
	})

	t.Run("fails to create tag without name", func(t *testing.T) {
		tag := Tag{Color: "#fecaca"}

		err := tag.Create(db)
		if err == nil {
			t.Error("Expected error when creating tag without name")
		}

		if err.Error() != "tag name is required" {
			t.Errorf("Expected 'tag name is required', got %s", err.Error())
		}
	})

	t.Run("fails to create duplicate tag name", func(t *testing.T) {
		tag1 := Tag{Name: "duplicate", Color: "#fecaca"}
		err := tag1.Create(db)
		if err != nil {
			t.Errorf("First tag creation failed: %v", err)
		}

		tag2 := Tag{Name: "duplicate", Color: "#fed7aa"}
		err = tag2.Create(db)
		if err == nil {
			t.Error("Expected error when creating tag with duplicate name")
		}
	})
}

func TestTagSave(t *testing.T) {
	db := setupTagTestDB(t)

	tag := Tag{Name: "test", Color: "#fecaca"}
	err := tag.Create(db)
	if err != nil {
		t.Fatalf("Failed to create initial tag: %v", err)
	}

	tag.Color = "#fed7aa"
	err = tag.Save(db)
	if err != nil {
		t.Errorf("Save returned error: %v", err)
	}

	var dbTag Tag
	result := db.First(&dbTag, "id = ?", tag.ID)
	if result.Error != nil {
		t.Errorf("Tag not found: %v", result.Error)
	}

	if dbTag.Color != "#fed7aa" {
		t.Errorf("Expected color '#fed7aa', got %s", dbTag.Color)
	}
}

func TestTagLoadByID(t *testing.T) {
	db := setupTagTestDB(t)

	originalTag := Tag{Name: "loadtest", Color: "#fecaca"}
	err := originalTag.Create(db)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	t.Run("loads existing tag", func(t *testing.T) {
		var loadedTag Tag
		err := loadedTag.LoadByID(db, originalTag.ID)
		if err != nil {
			t.Errorf("LoadByID returned error: %v", err)
		}

		if loadedTag.Name != "loadtest" {
			t.Errorf("Expected name 'loadtest', got %s", loadedTag.Name)
		}

		if loadedTag.Color != "#fecaca" {
			t.Errorf("Expected color '#fecaca', got %s", loadedTag.Color)
		}
	})

	t.Run("fails to load non-existent tag", func(t *testing.T) {
		var loadedTag Tag
		nonExistentID := uuid.New()
		err := loadedTag.LoadByID(db, nonExistentID)
		if err == nil {
			t.Error("Expected error when loading non-existent tag")
		}

		if err.Error() != "tag not found" {
			t.Errorf("Expected 'tag not found', got %s", err.Error())
		}
	})
}

func TestTagUpdate(t *testing.T) {
	db := setupTagTestDB(t)

	tag := Tag{Name: "original", Color: "#fecaca"}
	err := tag.Create(db)
	if err != nil {
		t.Fatalf("Failed to create tag: %v", err)
	}

	t.Run("updates name and color", func(t *testing.T) {
		updateData := Tag{Name: "updated", Color: "#fed7aa"}
		err := tag.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if tag.Name != "updated" {
			t.Errorf("Expected name 'updated', got %s", tag.Name)
		}

		if tag.Color != "#fed7aa" {
			t.Errorf("Expected color '#fed7aa', got %s", tag.Color)
		}
	})

	t.Run("updates only name", func(t *testing.T) {
		originalColor := tag.Color
		updateData := Tag{Name: "name-only"}
		err := tag.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if tag.Name != "name-only" {
			t.Errorf("Expected name 'name-only', got %s", tag.Name)
		}

		if tag.Color != originalColor {
			t.Errorf("Expected color to remain %s, got %s", originalColor, tag.Color)
		}
	})

	t.Run("updates only color", func(t *testing.T) {
		originalName := tag.Name
		updateData := Tag{Color: "#bbf7d0"}
		err := tag.Update(db, &updateData)
		if err != nil {
			t.Errorf("Update returned error: %v", err)
		}

		if tag.Name != originalName {
			t.Errorf("Expected name to remain %s, got %s", originalName, tag.Name)
		}

		if tag.Color != "#bbf7d0" {
			t.Errorf("Expected color '#bbf7d0', got %s", tag.Color)
		}
	})
}

func TestTagDelete(t *testing.T) {
	db := setupTagTestDB(t)

	t.Run("deletes existing tag", func(t *testing.T) {
		tag := Tag{Name: "delete-me", Color: "#fecaca"}
		err := tag.Create(db)
		if err != nil {
			t.Fatalf("Failed to create tag: %v", err)
		}

		err = tag.Delete(db)
		if err != nil {
			t.Errorf("Delete returned error: %v", err)
		}

		var dbTag Tag
		result := db.First(&dbTag, "id = ?", tag.ID)
		if result.Error == nil {
			t.Error("Expected tag to be deleted")
		}
	})

	t.Run("fails to delete non-existent tag", func(t *testing.T) {
		tag := Tag{ID: uuid.New()}
		err := tag.Delete(db)
		if err == nil {
			t.Error("Expected error when deleting non-existent tag")
		}

		if err.Error() != "tag not found" {
			t.Errorf("Expected 'tag not found', got %s", err.Error())
		}
	})
}

func TestGetTags(t *testing.T) {
	db := setupTagTestDB(t)

	tag1 := Tag{Name: "work", Color: "#fecaca"}
	tag2 := Tag{Name: "personal", Color: "#fed7aa"}
	tag3 := Tag{Name: "urgent", Color: "#fde68a"}

	db.Create(&tag1)
	db.Create(&tag2)
	db.Create(&tag3)

	t.Run("gets all tags", func(t *testing.T) {
		tags, err := GetTags(db, "")
		if err != nil {
			t.Errorf("GetTags returned error: %v", err)
		}

		if len(tags) != 3 {
			t.Errorf("Expected 3 tags, got %d", len(tags))
		}
	})

	t.Run("filters tags by name", func(t *testing.T) {
		tags, err := GetTags(db, "work")
		if err != nil {
			t.Errorf("GetTags returned error: %v", err)
		}

		if len(tags) != 1 {
			t.Errorf("Expected 1 tag with 'work' in name, got %d", len(tags))
		}

		if tags[0].Name != "work" {
			t.Errorf("Expected tag name 'work', got %s", tags[0].Name)
		}
	})

	t.Run("filters tags by partial name", func(t *testing.T) {
		tags, err := GetTags(db, "p")
		if err != nil {
			t.Errorf("GetTags returned error: %v", err)
		}

		if len(tags) != 1 {
			t.Errorf("Expected 1 tag with 'p' in name, got %d", len(tags))
		}

		if tags[0].Name != "personal" {
			t.Errorf("Expected tag name 'personal', got %s", tags[0].Name)
		}
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		tags, err := GetTags(db, "nonexistent")
		if err != nil {
			t.Errorf("GetTags returned error: %v", err)
		}

		if len(tags) != 0 {
			t.Errorf("Expected 0 tags, got %d", len(tags))
		}
	})
}

func TestTagColors(t *testing.T) {
	if len(tagColors) == 0 {
		t.Error("Expected tagColors to contain at least one color")
	}

	for i, color := range tagColors {
		if color == "" {
			t.Errorf("Expected tagColors[%d] to be non-empty", i)
		}

		if color[0] != '#' {
			t.Errorf("Expected tagColors[%d] to start with '#', got %s", i, color)
		}

		if len(color) != 7 {
			t.Errorf("Expected tagColors[%d] to be 7 characters long, got %d", i, len(color))
		}
	}
}
