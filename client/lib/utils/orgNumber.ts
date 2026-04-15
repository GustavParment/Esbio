// Swedish organisationsnummer helpers.
// Format: 10 digits as "XXXXXX-XXXX", with a Luhn checksum on all 10 digits.

export function digitsOnly(s: string): string {
  return (s || "").replace(/\D/g, "");
}

export function formatOrgNumber(input: string): string {
  const d = digitsOnly(input).slice(0, 10);
  if (d.length <= 6) return d;
  return `${d.slice(0, 6)}-${d.slice(6)}`;
}

// Returns null if valid, otherwise a Swedish error message. Empty is invalid —
// org number is mandatory throughout the app.
export function validateOrgNumber(input: string): string | null {
  const d = digitsOnly(input);
  if (d.length === 0) return "Organisationsnummer krävs";
  if (d.length !== 10) return "Organisationsnummer måste vara 10 siffror (t.ex. 556677-8899)";
  if (!luhnValid(d)) return "Ogiltigt organisationsnummer (kontrollsiffran stämmer inte)";
  return null;
}

function luhnValid(digits: string): boolean {
  let sum = 0;
  for (let i = 0; i < digits.length; i++) {
    let n = parseInt(digits[i], 10);
    if (i % 2 === 0) {
      n *= 2;
      if (n > 9) n -= 9;
    }
    sum += n;
  }
  return sum % 10 === 0;
}
