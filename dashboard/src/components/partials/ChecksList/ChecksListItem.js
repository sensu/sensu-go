import React from "react";
import PropTypes from "prop-types";
import gql from "graphql-tag";

import Menu from "@material-ui/core/Menu";
import MenuItem from "@material-ui/core/MenuItem";

import Code from "/components/Code";
import ListItem from "/components/partials/ListItem";

class CheckListItem extends React.Component {
  static propTypes = {
    check: PropTypes.object.isRequired,
    selected: PropTypes.bool.isRequired,
    onChangeSelected: PropTypes.func.isRequired,
    onClickSilence: PropTypes.func.isRequired,
  };

  static fragments = {
    check: gql`
      fragment ChecksListItem_check on CheckConfig {
        name
        command
        source
        subscriptions
        interval
      }
    `,
  };

  state = { menuOpen: false };

  _menuAnchorRef = React.createRef();

  openMenu = () => {
    this.setState({ menuOpen: true });
  };
  closeMenu = () => {
    this.setState({ menuOpen: false });
  };

  render() {
    const { check, selected, onChangeSelected, onClickSilence } = this.props;

    return (
      <ListItem
        selected={selected}
        onChangeSelected={onChangeSelected}
        title={<strong>{check.name}</strong>}
        details={
          <React.Fragment>
            <Code>{check.command}</Code>
            <br />
            Executed every{" "}
            <strong>
              {check.interval} {check.interval === 1 ? "second" : "seconds"}
            </strong>{" "}
            by{" "}
            <strong>
              {check.subscriptions.length}{" "}
              {check.subscriptions.length === 1
                ? "subscription"
                : "subscriptions"}
            </strong>.
          </React.Fragment>
        }
        renderMenu={({ anchorEl, close }) => (
          <Menu open onClose={close} anchorEl={anchorEl}>
            <MenuItem
              onClick={() => {
                onClickSilence();
                this.closeMenu();
              }}
            >
              Silence
            </MenuItem>
          </Menu>
        )}
      />
    );
  }
}

export default CheckListItem;
