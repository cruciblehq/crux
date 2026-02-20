-- Deletes a channel.
--
-- Does not delete the channel's version or archives, only the channel metadata.
DELETE FROM channels
WHERE namespace = ? AND resource = ? AND name = ?;
