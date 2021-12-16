INNER JOIN (SELECT i.image_id, public_url
			FROM files f
			JOIN images i ON i.image_id = f.image_id
			WHERE width = (
						SELECT MIN(width) 
						FROM files f1
						WHERE f1.image_id = f.image_id 
						AND f1.width > 400 OR (SELECT MAX(width) FROM files f2 WHERE f1.image_id = f2.image_id) < 400
						)
			) AS small_images 
ON small_images.image_id = cover_image_id
ORDER BY latest_date DESC