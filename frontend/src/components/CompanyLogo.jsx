import placeholder from '../assets/company-placeholder.svg';

// Logo filenames follow the company name, e.g. "Acme Labs" → acme-labs.png.
const logos = import.meta.glob('../assets/company-logos/*.{png,svg,webp,jpg,jpeg}', {
  eager: true,
  import: 'default',
  query: '?url',
});
const logoByCompany = Object.fromEntries(
  Object.entries(logos).map(([path, url]) => [path.split('/').pop().replace(/\.[^.]+$/, ''), url])
);

export function companyLogoSource(company, logoPath) {
  if (logoPath) return (import.meta.env.VITE_API_URL || '') + logoPath;
  const slug = (company ?? '').trim().toLowerCase().replace(/[^\p{L}\p{N}]+/gu, '-').replace(/^-|-$/g, '');
  return Object.hasOwn(logoByCompany, slug) ? logoByCompany[slug] : null;
}

export default function CompanyLogo({ company, logoPath }) {
  const src = companyLogoSource(company, logoPath) || placeholder;
  return (
    <div className="company-logo" aria-hidden="true">
      <img src={src} alt="" />
    </div>
  );
}
