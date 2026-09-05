/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
/** Abstract ice facets and wake lines, derived from the brand palette and geometry. */
export function StationLandscape() {
  return (
    <svg
      className='station-landscape'
      viewBox='0 0 1440 860'
      fill='none'
      preserveAspectRatio='xMidYMid slice'
      aria-hidden='true'
      focusable='false'
    >
      <path
        className='ice-field'
        d='M0 702C270 655 412 706 662 630C905 556 1120 612 1440 510V860H0Z'
      />
      <g className='ice-facets'>
        <path
          className='ice-face-light'
          d='M705 641 843 291 968 166 1034 371 1154 458 1100 709Z'
        />
        <path className='ice-face-mid' d='m705 641 138-350 14 235Z' />
        <path className='ice-face-white' d='m843 291 125-125-53 310Z' />
        <path className='ice-face-blue' d='m968 166 66 205-119 105Z' />
        <path className='ice-face-white' d='m857 526 58-50 119-105 66 338Z' />
        <path className='ice-face-mid' d='m1034 371 120 87-54 251Z' />
        <path
          className='ice-face-light'
          d='m1116 650 84-287 119-111 75 225 46 52v188Z'
        />
        <path className='ice-face-white' d='m1200 363 119-111-38 306Z' />
        <path className='ice-face-blue' d='m1319 252 75 225-113 81Z' />
        <path className='ice-face-mid' d='m1116 650 165-92 159-29v188Z' />
      </g>
      <path
        className='ice-shore'
        d='M0 730C251 671 459 776 719 699C1010 612 1182 708 1440 606V860H0Z'
      />
      <path
        className='ice-wake-back'
        d='M-70 846C120 733 619 822 776 738C919 661 543 704 583 650C614 609 1033 614 1176 551'
      />
      <path
        className='ice-wake'
        d='M-80 825C160 720 602 785 739 721C908 642 537 686 589 633C647 575 958 645 1136 564'
      />
      <path
        className='ice-route'
        d='M-70 783C200 685 559 753 668 696C749 654 563 646 616 611C678 570 1036 586 1175 520'
      />
      <path className='ice-crystal' d='m1129 218 18-31 25 16-15 31Z' />
      <path className='ice-crystal' d='m755 369 11-20 17 9-8 24Z' />
    </svg>
  )
}
