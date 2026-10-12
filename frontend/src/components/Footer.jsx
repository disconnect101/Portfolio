export default function Footer({ profile }) {
  return (
    <footer className="footer">
      <div className="container">
        <p>© {new Date().getFullYear()} {profile.name}. All rights reserved.</p>
      </div>
    </footer>
  );
}
