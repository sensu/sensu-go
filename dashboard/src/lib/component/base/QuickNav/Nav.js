import React from "react";
import PropTypes from "prop-types";

import CheckIcon from "/lib/component/icon/Check";
import EntityIcon from "/lib/component/icon/Entity";
import EventIcon from "/lib/component/icon/Event";
import SilenceIcon from "/lib/component/icon/Silence";

import Button from "./Button";

class QuickNav extends React.Component {
  static propTypes = {
    className: PropTypes.string,
    namespace: PropTypes.string.isRequired,
  };

  static defaultProps = { className: "" };

  render() {
    const { className, namespace } = this.props;

    return (
      <div className={className}>
        <Button
          namespace={namespace}
          Icon={EventIcon}
          caption="Events"
          to="events"
        />
        <Button
          namespace={namespace}
          Icon={EntityIcon}
          caption="Entities"
          to="entities"
        />
        <Button
          namespace={namespace}
          Icon={CheckIcon}
          caption="Checks"
          to="checks"
        />
        <Button
          namespace={namespace}
          Icon={SilenceIcon}
          caption="Silences"
          to="silences"
        />
      </div>
    );
  }
}

export default QuickNav;
