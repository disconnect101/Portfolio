export default function About({ profile }) {
  return (
    <section id="about" className="section">
      <div className="container narrow">
        <h2>About me</h2>
        <p className="lead">{profile.description}</p>
        <div className="links">
          {profile.githubUrl && <a className="btn" href={profile.githubUrl} target="_blank" rel="noreferrer">GitHub</a>}
          {profile.linkedinUrl && <a className="btn" href={profile.linkedinUrl} target="_blank" rel="noreferrer">LinkedIn</a>}
          {profile.email && <a className="btn primary" href={`mailto:${profile.email}`}>Get in touch</a>}
        </div>
      </div>
    </section>
  );
}
