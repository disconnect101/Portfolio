export default function About({ profile }) {
  return (
    <section id="about" className="about">
      <h2>About me</h2>
      <p className="lead">{profile.description}</p>
      <div className="contact-actions" aria-label="Contact links">
        {profile.githubUrl && (
          <a className="contact-link contact-link-social" href={profile.githubUrl} target="_blank" rel="noreferrer">
            <svg aria-hidden="true" viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.58 2 12.23c0 4.52 2.87 8.35 6.84 9.71.5.1.68-.22.68-.49 0-.24-.01-1.04-.01-1.88-2.78.62-3.37-1.21-3.37-1.21-.45-1.19-1.11-1.5-1.11-1.5-.91-.64.07-.63.07-.63 1 .07 1.53 1.06 1.53 1.06.9 1.57 2.35 1.12 2.92.86.09-.67.35-1.12.64-1.38-2.22-.26-4.56-1.15-4.56-5.12 0-1.13.39-2.05 1.03-2.78-.1-.26-.45-1.32.1-2.75 0 0 .84-.28 2.75 1.06A9.32 9.32 0 0 1 12 5.88c.85 0 1.7.12 2.5.35 1.9-1.34 2.74-1.06 2.74-1.06.55 1.43.2 2.49.1 2.75.64.73 1.03 1.65 1.03 2.78 0 3.98-2.34 4.85-4.57 5.11.36.33.68.96.68 1.94 0 1.4-.01 2.53-.01 2.88 0 .27.18.59.69.49A10.22 10.22 0 0 0 22 12.23C22 6.58 17.52 2 12 2Z" /></svg>
            <span>GitHub</span>
          </a>
        )}
        {profile.linkedinUrl && (
          <a className="contact-link contact-link-social" href={profile.linkedinUrl} target="_blank" rel="noreferrer">
            <svg aria-hidden="true" viewBox="0 0 24 24"><path d="M5.37 3.5a1.87 1.87 0 1 1-3.74 0 1.87 1.87 0 0 1 3.74 0ZM1.9 7.08h3.2V22H1.9V7.08ZM7.1 7.08h3.06v2.04h.04c.43-.8 1.47-2.04 3.03-2.04 3.24 0 3.84 2.2 3.84 5.05V22h-3.2v-8.76c0-2.1-.04-4.8-2.84-4.8-2.84 0-3.28 2.28-3.28 4.65V22H7.1V7.08Z" /></svg>
            <span>LinkedIn</span>
          </a>
        )}
        {profile.email && (
          <a className="contact-link contact-link-primary" href={`mailto:${profile.email}`}>
            <span>Get in touch</span>
            <svg aria-hidden="true" viewBox="0 0 24 24"><path d="M5 12h13M13 6l6 6-6 6" /></svg>
          </a>
        )}
      </div>
    </section>
  );
}
