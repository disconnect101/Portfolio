export default function Experience({ items }) {
  return (
    <section id="experience" className="section">
      <div className="container narrow">
        <h2>Experience</h2>
        <div className="timeline">
          {items.map((e) => (
            <div className="timeline-item" key={e.id}>
              <div className="timeline-date">{e.startDate} – {e.endDate}</div>
              <h3>{e.role}</h3>
              <div className="timeline-org">{e.company}{e.location ? ` · ${e.location}` : ''}</div>
              <ul>
                {e.highlights.map((h, i) => <li key={i}>{h}</li>)}
              </ul>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
