-- +goose Up
ALTER TABLE cv2x_messages ALTER COLUMN plmn_info TYPE VARCHAR(50);

-- +goose Down

