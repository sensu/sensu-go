import chalk from "chalk";

export const success = () => {
  const list = ["🎉", "🎊", "🔥", "🌮", "💎", "🌭", "🙌", "✅", "👌", "🧐"];
  return list[Math.floor(Math.random() * list.length)];
};

let interval;

export const loading = {
  start(message = "") {
    if (interval) {
      interval = clearInterval(interval);
      console.log();
    }

    process.stdout.write(`🕛 ${message}`);

    if (process.stdout.isTTY) {
      let time = 0;

      interval = setInterval(() => {
        const clock = String.fromCodePoint(128336 + time);

        process.stdout.write(`\r${clock} ${message}`);

        time = (time + 1) % 12;
      }, 33);
    }
  },
  stop: (abort = false) => {
    interval = clearInterval(interval);

    if (abort) {
      console.log();
    } else {
      console.log(process.stdout.isTTY ? `\r${success()}` : "");
      console.log(chalk.gray("   Done!"));
    }
  },
};

["SIGINT", "SIGTERM"].forEach(sig => {
  process.on(sig, () => loading.stop(true));
});
