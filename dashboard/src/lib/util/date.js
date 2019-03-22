export const parseUNIX = timestamp => new Date(timestamp * 1000);

const monthNameFormatter = new Intl.DateTimeFormat("en-US", { month: "long" });
export const getMonthName = date => monthNameFormatter.format(date);

const timeZoneNameFormatter = new Intl.DateTimeFormat("en-US", {
  month: "numeric",
  timeZoneName: "short",
});
export const getTimeZoneName = date => {
  const parts = timeZoneNameFormatter.formatToParts(date);

  for (let i = parts.length - 1; i >= 0; i -= 1) {
    if (parts[i].type === "timeZoneName") {
      return parts[i].value;
    }
  }

  return "";
};

const hourFormatter = new Intl.DateTimeFormat("en-us", { hour: "numeric" });
export const getHour = date => {
  const parts = hourFormatter.formatToParts(date);

  for (let i = 0; i < parts.length; i += 1) {
    if (parts[i].type === "hour") {
      return parts[i].value;
    }
  }

  return "";
};

export const getDayperiod = date => {
  const parts = hourFormatter.formatToParts(date);

  for (let i = parts.length - 1; i >= 0; i -= 1) {
    if (parts[i].type === "dayperiod" || parts[i].type === "dayPeriod") {
      return parts[i].value;
    }
  }

  return "";
};
