export default function Hero({ profile }) {
  return (
    <section id="top" className="hero">
      <div
        className="hero-cover"
        style={{ backgroundImage: `url(${profile.coverPhotoUrl})` }}
        role="img"
        aria-label="Cover photo"
      />
      <div className="container hero-content">
        <img className="avatar" src={profile.profilePhotoUrl} alt={profile.name} />
        <h1>{profile.name}</h1>
        <p className="hero-title">{profile.title}</p>
        {profile.location && <p className="hero-location">📍 {profile.location}</p>}
      </div>
    </section>
  );
}
