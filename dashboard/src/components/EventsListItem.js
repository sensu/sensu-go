import React from "react";
import PropTypes from "prop-types";
import moment from "moment";
import { compose } from "recompose";

import gql from "graphql-tag";
import { withApollo } from "react-apollo";
import { withStyles } from "material-ui/styles";
import Typography from "material-ui/Typography";
import Menu, { MenuItem } from "material-ui/Menu";
import Button from "material-ui/ButtonBase";

import Checkbox from "material-ui/Checkbox";
import Disclosure from "material-ui-icons/MoreVert";

import resolveEvent from "/mutations/resolveEvent";
import EventStatus from "/components/EventStatus";
import { TableListItem } from "/components/TableList";

const styles = theme => ({
  root: {
    color: theme.palette.text.secondary,
    "& strong": {
      fontWeight: "normal",
      color: theme.palette.text.primary,
    },
  },
  checkbox: {
    display: "inline-block",
    verticalAlign: "top",
    marginLeft: -12,
  },
  status: {
    display: "inline-block",
    verticalAlign: "top",
    padding: "14px 0",
  },
  disclosure: {
    color: theme.palette.action.active,
    marginLeft: 12,
    paddingTop: 14,
  },
  content: {
    width: "calc(100% - 96px)",
    display: "inline-block",
    padding: 14,
  },
  command: {
    fontSize: "0.8125rem",
    whiteSpace: "nowrap",
    overflow: "hidden",
    textOverflow: "ellipsis",
  },
  timeHolder: {
    width: "100%",
    display: "flex",
    fontSize: "0.8125rem",
    margin: "4px 0 6px",
  },
  pipe: { marginTop: -4 },
});

function fromNow(date) {
  const delta = new Date(date) - new Date();
  if (delta < 0) {
    return moment.duration(delta).humanize(true);
  }
  return "just now";
}

class EventListItem extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    checked: PropTypes.bool.isRequired,
    onChange: PropTypes.func.isRequired,
    client: PropTypes.object.isRequired,
    event: PropTypes.shape({
      entity: PropTypes.shape({
        name: PropTypes.string.isRequired,
      }).isRequired,
      check: PropTypes.shape({
        name: PropTypes.string.isRequired,
        output: PropTypes.string.isRequired,
      }).isRequired,
      timestamp: PropTypes.string.isRequired,
    }).isRequired,
  };

  static fragments = {
    event: gql`
      fragment EventsListItem_event on Event {
        ... on Event {
          id
          timestamp
          check {
            status
            name
            output
          }
          entity {
            name
          }
        }
      }
    `,
  };

  constructor(props) {
    super(props);
    this.fromNow = fromNow(props.event.timestamp);
  }

  state = { anchorEl: null };

  componentWillReceiveProps(nextProps) {
    if (this.props.event.timestamp !== nextProps.event.timestamp) {
      this.fromNow = fromNow(nextProps.event.timestamp);
    }
  }

  onClose = () => {
    this.setState({ anchorEl: null });
  };

  handleClick = event => {
    this.setState({ anchorEl: event.currentTarget });
  };

  silenceEntity = entity => () => {
    // eslint-disable-next-line no-console
    console.info("implement entity silencing!", "id:", entity);
    this.setState({ anchorEl: null });
  };

  silenceCheck = check => () => {
    // eslint-disable-next-line no-console
    console.info("implement check silencing!", "id:", check);
    this.setState({ anchorEl: null });
  };

  resolve = () => {
    const { client, event } = this.props;
    resolveEvent(client, event.id);
    this.setState({ anchorEl: null });
  };

  render() {
    const { checked, classes, event: { entity, check }, onChange } = this.props;
    const { anchorEl } = this.state;
    const time = this.fromNow;

    return (
      <TableListItem className={classes.root} selected={checked}>
        <div className={classes.checkbox}>
          <Checkbox color="primary" onChange={onChange} checked={checked} />
        </div>
        <div className={classes.status}>
          <EventStatus status={check.status} />
        </div>
        <div className={classes.content}>
          <span className={classes.caption}>
            <strong>
              {entity.name} › {check.name}
            </strong>
          </span>
          <div className={classes.timeHolder}>
            Last occurred <strong>&nbsp;{time}&nbsp;</strong> and exited with
            status <strong>&nbsp;{check.status}.</strong>
          </div>
          <Typography variant="caption" className={classes.command}>
            {check.output}
          </Typography>
        </div>
        <div className={classes.disclosure}>
          <Button onClick={this.handleClick}>
            <Disclosure />
          </Button>
          {/* TODO give these functionality, pass correct value */}
          <Menu
            anchorEl={anchorEl}
            open={Boolean(anchorEl)}
            onClose={this.onClose}
            id="silenceButton"
          >
            <MenuItem
              key={"silence-Entity"}
              onClick={this.silenceEntity("entity")}
            >
              Silence Entity
            </MenuItem>
            <MenuItem
              key={"silence-Check"}
              onClick={this.silenceCheck("entity")}
            >
              Silence Check
            </MenuItem>
            <MenuItem key={"resolve"} onClick={this.resolve}>
              Resolve
            </MenuItem>
          </Menu>
        </div>
      </TableListItem>
    );
  }
}

export default compose(withStyles(styles), withApollo)(EventListItem);
