
export const titleDirective = (title) => {
    let titleWords  = title.split('_').slice(1);
    titleWords = titleWords.map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    return titleWords.join(' ');
  } 