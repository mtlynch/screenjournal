export function tzOffsetString() {
  const offsetHours = -(new Date().getTimezoneOffset() / 60);
  let offsetString = "";
  if (offsetHours >= 0) {
    offsetString += "+";
  } else {
    offsetString += "-";
  }
  if (Math.abs(offsetHours) <= 10) {
    offsetString += "0";
  }
  offsetString += "" + Math.abs(Math.floor(offsetHours));
  offsetString += ":";
  const tzOffsetMins = (offsetHours - Math.floor(offsetHours)) * 10;
  offsetString += tzOffsetMins;
  if (tzOffsetMins < 10) {
    offsetString += "0";
  }
  return offsetString;
}
