import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";
import { withApollo } from "react-apollo";

import Checkbox from "@material-ui/core/Checkbox";
import IconButton from "@material-ui/core/IconButton";
import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";
import MoreVert from "@material-ui/icons/MoreVert";
import RootRef from "@material-ui/core/RootRef";
import TableCell from "@material-ui/core/TableCell";

import Code from "/components/Code";
import resolveEvent from "/mutations/resolveEvent";
import RelativeDate from "/components/RelativeDate";

import MenuController from "/components/controller/MenuController";

import ResourceDetails from "/components/partials/ResourceDetails";
import TableOverflowCell from "/components/partials/TableOverflowCell";
import TableSelectableRow from "/components/partials/TableSelectableRow";

import NamespaceLink from "/components/util/NamespaceLink";
import CheckStatusIcon from "/components/CheckStatusIcon";

class EventListItem extends React.Component {
  static propTypes = {
    selected: PropTypes.bool.isRequired,
    onChangeSelected: PropTypes.func.isRequired,
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
    }).isRequired,
  };

  static fragments = {
    event: gql`
      fragment EventsListItem_event on Event {
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
    `,
  };

  resolve = () => {
    const { client, event } = this.props;
    resolveEvent(client, event);
  };

  renderMenu = ({ close, anchorEl }) => {
    const { event } = this.props;

    return (
      <Menu open onClose={close} anchorEl={anchorEl}>
        <MenuItem
          key={"silence-Entity"}
          onClick={() => {
            this.props.onClickSilenceEntity();
            close();
          }}
        >
          Silence Entity
        </MenuItem>
        <MenuItem
          key={"silence-Check"}
          onClick={() => {
            this.props.onClickSilenceCheck();
            close();
          }}
        >
          Silence Check
        </MenuItem>
        {event.check &&
          event.check.status !== 0 && (
            <MenuItem
              onClick={() => {
                this.resolve();
                close();
              }}
            >
              Resolve
            </MenuItem>
          )}
      </Menu>
    );
  };

  render() {
    const { selected, event, onChangeSelected } = this.props;
    const { entity, check, timestamp } = event;

    return (
      <TableSelectableRow selected={selected}>
        <TableCell padding="checkbox">
          <Checkbox
            color="primary"
            checked={selected}
            onChange={e => onChangeSelected(e.target.checked)}
          />
        </TableCell>
        <TableOverflowCell>
          <ResourceDetails
            icon={
              event.check && <CheckStatusIcon statusCode={event.check.status} />
            }
            title={
              <NamespaceLink
                namespace={event.namespace}
                to={`/events/${entity.name}/${check.name}`}
              >
                {entity.name} › {check.name}
              </NamespaceLink>
            }
            details={
              <React.Fragment>
                Last occurred{" "}
                <strong>
                  <RelativeDate dateTime={timestamp} />
                </strong>{" "}
                and exited with status <strong>{check.status}</strong>.
                {check.output && (
                  <React.Fragment>
                    <br />
                    <Code>{check.output}</Code>
                  </React.Fragment>
                )}
              </React.Fragment>
            }
            renderMenu={this.renderMenu}
          />
        </TableOverflowCell>
        <TableCell padding="checkbox">
          <MenuController renderMenu={this.renderMenu}>
            {({ open, ref }) => (
              <RootRef rootRef={ref}>
                <IconButton onClick={open}>
                  <MoreVert />
                </IconButton>
              </RootRef>
            )}
          </MenuController>
        </TableCell>
      </TableSelectableRow>
    );
  }
}

export default withApollo(EventListItem);
