import About from './About.jsx';

export default function Hero({ profile }) {
  return (
    <section id="top" className="hero container" aria-labelledby="profile-name">
      <div className="hero-cover">
        {profile.coverPhotoUrl && <img src={profile.coverPhotoUrl} alt="" />}
      </div>
      <div className="intro">
        <header className="profile-summary">
          {profile.profilePhotoUrl && <img className="avatar" src={profile.profilePhotoUrl} alt="" />}
          <h1 id="profile-name">{profile.name}</h1>
          <p className="hero-title">{profile.title}</p>
          {profile.location && <p className="hero-location">{profile.location}</p>}
        </header>
        <About profile={profile} />
      </div>
    </section>
  );
}
