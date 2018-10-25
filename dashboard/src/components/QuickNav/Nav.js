import React from "react";
import PropTypes from "prop-types";

import CheckIcon from "/icons/Check";
import EntityIcon from "/icons/Entity";
import EventIcon from "/icons/Event";
import SilenceIcon from "/icons/Silence";

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
          caption="Silenced"
          to="silences"
        />
      </div>
    );
  }
}

export default QuickNav;
