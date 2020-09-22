/**
 *
 * @param {string} title resource name
 * @returns title without prefix(aws_) and uppercase first letter foreach word
 */
export const titleDirective = (title) => {
  if (!title || title.indexOf("_") === -1) {
    return title;
  }
  let titleWords = title.split("_").slice(1);
  titleWords = titleWords.map(
    (word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase()
  );
  return titleWords.join(" ");
};

export const ucfirstDirective = (word) =>
  word.charAt(0).toUpperCase() + word.slice(1).toLowerCase();
