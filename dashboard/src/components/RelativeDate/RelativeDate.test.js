import React from "react";
import { mount } from "enzyme";
import IntlRelativeFormat from "intl-relativeformat";
import RelativeDate from "./RelativeDate";

global.IntlRelativeFormat = IntlRelativeFormat;

test("RelativeDate given date seconds ago", () => {
  const now = new Date();
  const date = new Date(now.getTime() - 12000);
  const elem = mount(<RelativeDate to={now} dateTime={date.toUTCString()} />);

  expect(elem.text()).toEqual("seconds ago");
});

test("RelativeDate given date seconds ahead", () => {
  const now = new Date();
  const date = new Date(now.getTime() + 6000);
  const elem = mount(<RelativeDate to={now} dateTime={date.toUTCString()} />);

  expect(elem.text()).toEqual("in a few seconds");
});

test("RelativeDate given date that will occur in less than a minute", () => {
  const now = new Date();
  const date = new Date(now.getTime() + 20000);
  const elem = mount(<RelativeDate to={now} dateTime={date.toUTCString()} />);

  expect(elem.text()).toEqual("in less than a minute");
});

test("RelativeDate given capitalize prop", () => {
  const now = new Date();
  const date = new Date(now.getTime() + 20000);
  const elem = mount(
    <RelativeDate to={now} dateTime={date.toUTCString()} capitalize />,
  );

  expect(elem.text()).toEqual("In less than a minute");
});
