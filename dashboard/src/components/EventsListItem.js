import React from "react";
import PropTypes from "prop-types";
import { compose } from "recompose";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";
import { withStyles } from "@material-ui/core/styles";
import Typography from "@material-ui/core/Typography";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";

import resolveEvent from "/mutations/resolveEvent";
import RelativeDate from "/components/RelativeDate";
import StatusListItem from "/components/StatusListItem";
import NamespaceLink from "/components/util/NamespaceLink";

const styles = () => ({
  command: {
    whiteSpace: "nowrap",
    overflow: "hidden",
    textOverflow: "ellipsis",
  },
  timeHolder: {
    marginBottom: 6,
  },
});

class EventListItem extends React.Component {
  static propTypes = {
    classes: PropTypes.object.isRequired,
    checked: PropTypes.bool.isRequired,
    onClickSelect: PropTypes.func.isRequired,
    onClickSilenceEntity: PropTypes.func.isRequired,
    onClickSilenceCheck: PropTypes.func.isRequired,
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
      deleted: PropTypes.bool.isRequired,
    }).isRequired,
  };

  static fragments = {
    event: gql`
      fragment EventsListItem_event on Event {
        ... on Event {
          id
          timestamp
          deleted @client
          check {
            status
            name
            output
          }
          entity {
            name
          }
          namespace {
            organization
            environment
          }
        }
      }
    `,
  };

  resolve = () => {
    const { client, event } = this.props;
    resolveEvent(client, event);
  };

  renderMenu = ({ open, onClose, anchorEl }) => {
    const { event } = this.props;

    return (
      <Menu open={open} onClose={onClose} anchorEl={anchorEl}>
        <MenuItem
          key={"silence-Entity"}
          onClick={() => {
            this.props.onClickSilenceEntity();
            onClose();
          }}
        >
          Silence Entity
        </MenuItem>
        <MenuItem
          key={"silence-Check"}
          onClick={() => {
            this.props.onClickSilenceCheck();
            onClose();
          }}
        >
          Silence Check
        </MenuItem>
        {event.check &&
          event.check.status !== 0 && (
            <MenuItem
              onClick={() => {
                this.resolve();
                onClose();
              }}
            >
              Resolve
            </MenuItem>
          )}
      </Menu>
    );
  };

  render() {
    const { checked, classes, event, onClickSelect } = this.props;
    const { entity, check, timestamp } = event;

    // NOTE: Replace this when we add pagination to lists.
    if (event.deleted) {
      return null;
    }

    return (
      <StatusListItem
        selected={checked}
        onClickSelect={onClickSelect}
        status={event.check && event.check.status}
        title={
          <NamespaceLink
            namespace={event.namespace}
            to={`/events/${entity.name}/${check.name}`}
          >
            <strong>
              {entity.name} › {check.name}
            </strong>
          </NamespaceLink>
        }
        renderMenu={this.renderMenu}
      >
        <div className={classes.timeHolder}>
          <p>
            Last occurred{" "}
            <strong>
              <RelativeDate dateTime={timestamp} />
            </strong>{" "}
            and exited with status <strong>{check.status}</strong>.
          </p>
        </div>
        <Typography
          component="div"
          variant="caption"
          className={classes.command}
        >
          {check.output || <span>&nbsp;</span>}
        </Typography>
      </StatusListItem>
    );
  }
}

export default compose(withStyles(styles), withApollo)(EventListItem);
