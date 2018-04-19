import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";
import { withStyles } from "material-ui/styles";

import { statusToColor } from "../../utils/checkStatus";
import OKIcon from "../../icons/SmallCheck";
import WarnIcon from "../../icons/WarnHollow";
import ErrIcon from "../../icons/ErrorHollow";

const styles = theme => ({
  inline: {
    fontSize: "inherit",
    verticalAlign: "middle",
  },
  green: {
    color: theme.palette.green,
  },
  yellow: {
    color: theme.palette.yellow,
  },
  orange: {
    color: theme.palette.orange,
  },
  red: {
    color: theme.palette.red,
  },
  muted: {
    color: theme.palette.grey[500],
  },
});

const componentMap = {
  ok: OKIcon,
  warning: WarnIcon,
  error: ErrIcon,
};

class Icon extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    status: PropTypes.oneOf(["ok", "warning", "error"]).isRequired,
    mutedOK: PropTypes.bool,
    inline: PropTypes.bool,
  };

  static defaultProps = {
    className: "",
    inline: false,
    mutedOK: false,
  };

  render() {
    const {
      classes,
      className: classNameProp,
      inline,
      mutedOK,
      status,
      ...props
    } = this.props;

    const Component = componentMap[status];
    const color = statusToColor(status);
    const className = classnames(classNameProp, classes[color], {
      [classes.muted]: status === "ok" && mutedOK,
      [classes.inline]: inline,
    });

    return <Component className={className} {...props} />;
  }
}

export default withStyles(styles)(Icon);
