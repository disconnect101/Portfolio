export default function Projects({ projects }) {
  return (
    <section id="projects" className="section alt">
      <div className="container">
        <h2>Projects</h2>
        <div className="grid">
          {projects.map((p) => (
            <article className="card" key={p.id}>
              {p.imageUrl && <img className="card-img" src={p.imageUrl} alt={p.title} />}
              <h3>{p.title}</h3>
              <p>{p.description}</p>
              <ul className="tags">
                {p.technologies.map((t) => <li key={t}>{t}</li>)}
              </ul>
              <div className="card-links">
                {p.repoUrl && <a href={p.repoUrl} target="_blank" rel="noreferrer">Code →</a>}
                {p.liveUrl && <a href={p.liveUrl} target="_blank" rel="noreferrer">Live →</a>}
              </div>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
