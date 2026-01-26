document.addEventListener("DOMContentLoaded", function () {
  const navLinks = document.querySelectorAll("nav a");
  const currentPath = window.location.pathname;

  navLinks.forEach((link) => {
    const href = link.getAttribute("href");
    // Handle both root path and .html pages
    if (
      currentPath === href ||
      (currentPath === "/" && href === "/") ||
      currentPath.endsWith(href)
    ) {
      link.classList.add("active");
    }
  });
});
