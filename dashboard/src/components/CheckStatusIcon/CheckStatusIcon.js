import React from "react";
import PropTypes from "prop-types";
import classnames from "classnames";

import { withStyles } from "@material-ui/core/styles";
import { statusCodeToId } from "/utils/checkStatus";

import SvgIcon from "@material-ui/core/SvgIcon";
import ErrIcon from "/icons/Error";
import ErrIconSm from "/icons/ErrorHollow";
import OKIcon from "/icons/OK";
import OKIconSm from "/icons/SmallCheck";
import WarnIcon from "/icons/Warn";
import WarnIconSm from "/icons/WarnHollow";
import UnknownIcon from "/icons/Unknown";
import SilenceIcon from "/icons/Silence";
import Tooltip from "@material-ui/core/Tooltip";

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
  silenced: {
    opacity: 0.35,
  },
  silenceIcon: {
    opacity: 0.71,
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
    silenced: PropTypes.bool,
  };

  static defaultProps = {
    className: "",
    inline: false,
    mutedOK: false,
    small: false,
    silenced: false,
  };

  render() {
    const {
      classes,
      className: classNameProp,
      inline,
      mutedOK,
      small,
      statusCode,
      silenced,

      ...props
    } = this.props;

    const status = statusCodeToId(statusCode);
    const Component = componentMap[!small ? "normal" : "small"][status];
    const className = classnames(classNameProp, classes[status], {
      [classes.muted]: status === "success" && mutedOK,
      [classes.inline]: inline,
      [classes.silenced]: silenced && !small,
    });

    const title = silenced ? "silenced" : status;
    const icon = <Component className={className} {...props} />;
    if (silenced) {
      if (small) {
        return (
          <Tooltip title={title}>
            <SilenceIcon className={className} />
          </Tooltip>
        );
      }
      return (
        <Tooltip title={title}>
          <SvgIcon viewBox="0 0 24 24">
            <SilenceIcon
              x={12}
              y={12}
              width={12}
              height={12}
              className={classes.silencedIcon}
            />
            {React.cloneElement(icon, {
              x: 0,
              y: 0,
              width: 24,
              height: 24,
              withGap: true,
            })}
          </SvgIcon>
        </Tooltip>
      );
    }

    return <Tooltip title={title}>{icon}</Tooltip>;
  }
}

export default withStyles(styles)(Icon);
