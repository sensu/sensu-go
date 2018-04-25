import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";

import { withStyles } from "material-ui/styles";
import OKIcon from "material-ui-icons/CheckCircle";
import WarnIcon from "material-ui-icons/Warning";
import ErrIcon from "material-ui-icons/Error";
import UnknownIcon from "material-ui-icons/Help";

import { statusCodeToId } from "/utils/checkStatus";
import OKIconSm from "/icons/SmallCheck";
import WarnIconSm from "/icons/WarnHollow";
import ErrIconSm from "/icons/ErrorHollow";

const styles = theme => ({
  inline: {
    fontSize: "inherit",
    verticalAlign: "middle",
  },
  success: {
    color: theme.palette.success,
  },
  warning: {
    color: theme.palette.warning,
  },
  critical: {
    color: theme.palette.critical,
  },
  unknown: {
    color: theme.palette.unknown,
  },
  muted: {
    color: theme.palette.grey[500],
  },
});

const componentMap = {
  normal: {
    success: OKIcon,
    warning: WarnIcon,
    critical: ErrIcon,
    unknown: UnknownIcon,
  },
  small: {
    success: OKIconSm,
    warning: WarnIconSm,
    critical: ErrIconSm,
    unknown: ErrIconSm,
  },
};

class Icon extends React.PureComponent {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    className: PropTypes.string,
    mutedOK: PropTypes.bool,
    small: PropTypes.bool,
    statusCode: PropTypes.number.isRequired,
    inline: PropTypes.bool,
  };

  static defaultProps = {
    className: "",
    inline: false,
    mutedOK: false,
    small: false,
  };

  render() {
    const {
      classes,
      className: classNameProp,
      inline,
      mutedOK,
      small,
      statusCode,

      ...props
    } = this.props;

    const status = statusCodeToId(statusCode);
    const Component = componentMap[!small ? "normal" : "small"][status];
    const className = classnames(classNameProp, classes[status], {
      [classes.muted]: status === "success" && mutedOK,
      [classes.inline]: inline,
    });

    return <Component className={className} {...props} />;
  }
}

export default withStyles(styles)(Icon);
