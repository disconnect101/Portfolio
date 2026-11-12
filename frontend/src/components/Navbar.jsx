export default function Navbar({ name }) {
  return (
    <header className="navbar">
      <div className="container navbar-inner">
        <a href="#top" className="brand">{name}</a>
        <nav>
          <a href="#about">About</a>
          <a href="#experience">Experience</a>
          <a href="#education">Education</a>
          <a href="#projects">Projects</a>
        </nav>
      </div>
    </header>
  );
}
