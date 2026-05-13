# Low-Level Design (LLD) Questions with Answers

## 1. Design a Parking Lot System

### Requirements
- Multiple floors, spots for different vehicle types
- Entry/exit gates with ticket generation
- Payment calculation based on duration

### Class Diagram

```java
// Enums
enum VehicleType { CAR, BIKE, TRUCK, VAN }
enum SpotType { COMPACT, LARGE, HANDICAPPED, BIKE }
enum PaymentStatus { PENDING, COMPLETED, FAILED }

// Main Classes
class ParkingLot {
    private String id;
    private List<Floor> floors;
    private List<Gate> entryGates;
    private List<Gate> exitGates;
    private static ParkingLot instance; // Singleton
    
    public static ParkingLot getInstance() {
        if (instance == null) {
            synchronized(ParkingLot.class) {
                if (instance == null) {
                    instance = new ParkingLot();
                }
            }
        }
        return instance;
    }
    
    public Ticket issueTicket(Vehicle vehicle, Gate gate) {
        ParkingSpot spot = findAvailableSpot(vehicle.getType());
        if (spot == null) throw new NoSpotAvailableException();
        
        spot.assignVehicle(vehicle);
        return new Ticket(vehicle, spot, gate, LocalDateTime.now());
    }
    
    public Payment processExit(Ticket ticket, Gate gate) {
        ParkingSpot spot = ticket.getSpot();
        spot.removeVehicle();
        
        double amount = calculateFee(ticket);
        return new Payment(ticket, amount, gate);
    }
    
    private ParkingSpot findAvailableSpot(VehicleType type) {
        for (Floor floor : floors) {
            ParkingSpot spot = floor.findSpot(type);
            if (spot != null) return spot;
        }
        return null;
    }
    
    private double calculateFee(Ticket ticket) {
        long hours = ChronoUnit.HOURS.between(
            ticket.getEntryTime(), 
            LocalDateTime.now()
        );
        return hours * ticket.getSpot().getHourlyRate();
    }
}

class Floor {
    private int floorNumber;
    private List<ParkingSpot> spots;
    
    public ParkingSpot findSpot(VehicleType vehicleType) {
        SpotType requiredSpotType = mapVehicleToSpot(vehicleType);
        return spots.stream()
            .filter(s -> s.getType() == requiredSpotType && s.isAvailable())
            .findFirst()
            .orElse(null);
    }
    
    private SpotType mapVehicleToSpot(VehicleType vType) {
        switch(vType) {
            case BIKE: return SpotType.BIKE;
            case CAR: return SpotType.COMPACT;
            case TRUCK: return SpotType.LARGE;
            default: return SpotType.COMPACT;
        }
    }
}

class ParkingSpot {
    private String id;
    private SpotType type;
    private boolean isAvailable;
    private Vehicle currentVehicle;
    private double hourlyRate;
    
    public synchronized void assignVehicle(Vehicle vehicle) {
        if (!isAvailable) throw new SpotOccupiedException();
        this.currentVehicle = vehicle;
        this.isAvailable = false;
    }
    
    public synchronized void removeVehicle() {
        this.currentVehicle = null;
        this.isAvailable = true;
    }
}

class Vehicle {
    private String licensePlate;
    private VehicleType type;
}

class Ticket {
    private String ticketId;
    private Vehicle vehicle;
    private ParkingSpot spot;
    private Gate entryGate;
    private LocalDateTime entryTime;
    
    public Ticket(Vehicle v, ParkingSpot s, Gate g, LocalDateTime time) {
        this.ticketId = UUID.randomUUID().toString();
        this.vehicle = v;
        this.spot = s;
        this.entryGate = g;
        this.entryTime = time;
    }
}

class Payment {
    private String paymentId;
    private Ticket ticket;
    private double amount;
    private PaymentStatus status;
    private LocalDateTime paymentTime;
    
    public boolean processPayment(PaymentMethod method) {
        // Integrate with payment gateway
        this.status = PaymentStatus.COMPLETED;
        return true;
    }
}

class Gate {
    private String gateId;
    private String type; // "ENTRY" or "EXIT"
}
```

### Key Design Patterns Used
- **Singleton**: ParkingLot instance
- **Factory**: Vehicle creation
- **Strategy**: Different payment methods
- **Observer**: Notify when spots become available

---

## 2. Design an LRU Cache

### Requirements
- O(1) get and put operations
- Fixed capacity, evict least recently used on overflow
- Thread-safe

### Implementation

```java
class LRUCache<K, V> {
    private final int capacity;
    private final Map<K, Node<K, V>> cache;
    private final DoublyLinkedList<K, V> list;
    
    public LRUCache(int capacity) {
        this.capacity = capacity;
        this.cache = new HashMap<>();
        this.list = new DoublyLinkedList<>();
    }
    
    public synchronized V get(K key) {
        if (!cache.containsKey(key)) {
            return null;
        }
        
        Node<K, V> node = cache.get(key);
        list.moveToHead(node);
        return node.value;
    }
    
    public synchronized void put(K key, V value) {
        if (cache.containsKey(key)) {
            Node<K, V> node = cache.get(key);
            node.value = value;
            list.moveToHead(node);
        } else {
            if (cache.size() >= capacity) {
                Node<K, V> tail = list.removeTail();
                cache.remove(tail.key);
            }
            
            Node<K, V> newNode = new Node<>(key, value);
            list.addToHead(newNode);
            cache.put(key, newNode);
        }
    }
    
    // Inner classes
    private static class Node<K, V> {
        K key;
        V value;
        Node<K, V> prev;
        Node<K, V> next;
        
        Node(K key, V value) {
            this.key = key;
            this.value = value;
        }
    }
    
    private static class DoublyLinkedList<K, V> {
        private Node<K, V> head;
        private Node<K, V> tail;
        
        DoublyLinkedList() {
            head = new Node<>(null, null);
            tail = new Node<>(null, null);
            head.next = tail;
            tail.prev = head;
        }
        
        void addToHead(Node<K, V> node) {
            node.next = head.next;
            node.prev = head;
            head.next.prev = node;
            head.next = node;
        }
        
        void removeNode(Node<K, V> node) {
            node.prev.next = node.next;
            node.next.prev = node.prev;
        }
        
        void moveToHead(Node<K, V> node) {
            removeNode(node);
            addToHead(node);
        }
        
        Node<K, V> removeTail() {
            Node<K, V> node = tail.prev;
            removeNode(node);
            return node;
        }
    }
}
```

### Time Complexity
- `get()`: O(1) - HashMap lookup + linked list operation
- `put()`: O(1) - HashMap insert + linked list operation

### Space Complexity
- O(capacity) - HashMap + Doubly Linked List



---

## 3. Design a Library Management System

### Requirements
- Manage books, members, librarians
- Issue/return books, track due dates
- Search books by title, author, ISBN
- Fine calculation for late returns

### Class Diagram

```java
// Enums
enum BookStatus { AVAILABLE, ISSUED, RESERVED, LOST }
enum MembershipType { BASIC, PREMIUM, STUDENT }
enum ReservationStatus { PENDING, COMPLETED, CANCELLED }

// Abstract Classes
abstract class Person {
    protected String id;
    protected String name;
    protected String email;
    protected String phone;
}

abstract class Account extends Person {
    protected String username;
    protected String password;
    protected AccountStatus status;
    
    public abstract boolean resetPassword();
}

// Main Classes
class Library {
    private String name;
    private Address location;
    private List<BookItem> books;
    private Map<String, Member> members;
    private Map<String, Librarian> librarians;
    
    public List<BookItem> searchByTitle(String title) {
        return books.stream()
            .filter(b -> b.getBook().getTitle().contains(title))
            .collect(Collectors.toList());
    }
    
    public List<BookItem> searchByAuthor(String author) {
        return books.stream()
            .filter(b -> b.getBook().getAuthors().contains(author))
            .collect(Collectors.toList());
    }
}

class Book {
    private String ISBN;
    private String title;
    private List<String> authors;
    private String publisher;
    private int publicationYear;
    private String category;
}

class BookItem {
    private String barcode;
    private Book book;
    private BookStatus status;
    private LocalDate dateOfPurchase;
    private double price;
    private Rack rack;
    
    public boolean checkout(Member member) {
        if (status != BookStatus.AVAILABLE) {
            return false;
        }
        
        BookLending lending = new BookLending(
            this, 
            member, 
            LocalDate.now(),
            LocalDate.now().plusDays(14)
        );
        
        this.status = BookStatus.ISSUED;
        member.addLending(lending);
        return true;
    }
    
    public boolean returnBook() {
        this.status = BookStatus.AVAILABLE;
        return true;
    }
}

class Member extends Account {
    private MembershipType type;
    private LocalDate dateOfMembership;
    private int totalBooksCheckedOut;
    private List<BookLending> currentLendings;
    private List<BookReservation> reservations;
    
    public boolean checkoutBook(BookItem book) {
        if (totalBooksCheckedOut >= getMaxBooksLimit()) {
            throw new MaxBooksLimitException();
        }
        
        if (book.checkout(this)) {
            totalBooksCheckedOut++;
            return true;
        }
        return false;
    }
    
    public boolean returnBook(BookItem book) {
        BookLending lending = findLending(book);
        if (lending == null) return false;
        
        lending.setReturnDate(LocalDate.now());
        
        // Calculate fine if overdue
        if (LocalDate.now().isAfter(lending.getDueDate())) {
            double fine = calculateFine(lending);
            lending.setFine(fine);
        }
        
        book.returnBook();
        totalBooksCheckedOut--;
        currentLendings.remove(lending);
        return true;
    }
    
    public boolean reserveBook(BookItem book) {
        BookReservation reservation = new BookReservation(
            this, 
            book, 
            LocalDate.now()
        );
        reservations.add(reservation);
        return true;
    }
    
    private int getMaxBooksLimit() {
        switch(type) {
            case BASIC: return 3;
            case PREMIUM: return 10;
            case STUDENT: return 5;
            default: return 3;
        }
    }
    
    private double calculateFine(BookLending lending) {
        long daysLate = ChronoUnit.DAYS.between(
            lending.getDueDate(), 
            LocalDate.now()
        );
        return daysLate * 1.0; // $1 per day
    }
    
    private BookLending findLending(BookItem book) {
        return currentLendings.stream()
            .filter(l -> l.getBookItem().equals(book))
            .findFirst()
            .orElse(null);
    }
}

class Librarian extends Account {
    public boolean addBookItem(BookItem book) {
        // Add book to library catalog
        return true;
    }
    
    public boolean blockMember(Member member) {
        member.setStatus(AccountStatus.BLOCKED);
        return true;
    }
    
    public boolean unblockMember(Member member) {
        member.setStatus(AccountStatus.ACTIVE);
        return true;
    }
}

class BookLending {
    private String lendingId;
    private BookItem bookItem;
    private Member member;
    private LocalDate issueDate;
    private LocalDate dueDate;
    private LocalDate returnDate;
    private double fine;
    
    public BookLending(BookItem book, Member member, 
                       LocalDate issue, LocalDate due) {
        this.lendingId = UUID.randomUUID().toString();
        this.bookItem = book;
        this.member = member;
        this.issueDate = issue;
        this.dueDate = due;
        this.fine = 0.0;
    }
}

class BookReservation {
    private String reservationId;
    private Member member;
    private BookItem bookItem;
    private LocalDate reservationDate;
    private ReservationStatus status;
    
    public BookReservation(Member m, BookItem b, LocalDate date) {
        this.reservationId = UUID.randomUUID().toString();
        this.member = m;
        this.bookItem = b;
        this.reservationDate = date;
        this.status = ReservationStatus.PENDING;
    }
}

class Rack {
    private String rackNumber;
    private String locationIdentifier;
}

class Fine {
    private double amount;
    private LocalDate creationDate;
    private Member member;
    
    public boolean collectFine(PaymentMethod method) {
        // Process payment
        return true;
    }
}

// Supporting Classes
class Address {
    private String street;
    private String city;
    private String state;
    private String zipCode;
    private String country;
}

enum AccountStatus { ACTIVE, BLOCKED, CLOSED }
```

### Key Features
1. **Search**: Multiple search strategies (title, author, ISBN)
2. **Checkout/Return**: Track lending with due dates
3. **Reservations**: Queue system for popular books
4. **Fines**: Automatic calculation for overdue books
5. **Member Types**: Different borrowing limits

### Design Patterns
- **Factory**: Create different member types
- **Strategy**: Different search algorithms
- **Observer**: Notify members when reserved books available
- **Singleton**: Library catalog

---

## 4. Design an Elevator System

### Requirements
- Multiple elevators in a building
- Efficient request handling
- Direction-based scheduling

### Implementation

```java
enum Direction { UP, DOWN, IDLE }
enum ElevatorStatus { MOVING, STOPPED, MAINTENANCE }

class ElevatorSystem {
    private List<Elevator> elevators;
    private ElevatorScheduler scheduler;
    
    public ElevatorSystem(int numElevators, int numFloors) {
        this.elevators = new ArrayList<>();
        for (int i = 0; i < numElevators; i++) {
            elevators.add(new Elevator(i, numFloors));
        }
        this.scheduler = new ElevatorScheduler(elevators);
    }
    
    public void requestElevator(int floor, Direction direction) {
        ExternalRequest request = new ExternalRequest(floor, direction);
        scheduler.scheduleRequest(request);
    }
}

class Elevator {
    private int id;
    private int currentFloor;
    private Direction direction;
    private ElevatorStatus status;
    private int capacity;
    private int currentLoad;
    private Queue<InternalRequest> requests;
    
    public Elevator(int id, int maxFloor) {
        this.id = id;
        this.currentFloor = 0;
        this.direction = Direction.IDLE;
        this.status = ElevatorStatus.STOPPED;
        this.capacity = 10;
        this.currentLoad = 0;
        this.requests = new PriorityQueue<>(new RequestComparator());
    }
    
    public void move() {
        if (requests.isEmpty()) {
            direction = Direction.IDLE;
            return;
        }
        
        InternalRequest nextRequest = requests.peek();
        
        if (nextRequest.getFloor() > currentFloor) {
            direction = Direction.UP;
            currentFloor++;
        } else if (nextRequest.getFloor() < currentFloor) {
            direction = Direction.DOWN;
            currentFloor--;
        } else {
            // Reached destination
            requests.poll();
            openDoors();
            closeDoors();
        }
    }
    
    public void addRequest(InternalRequest request) {
        requests.offer(request);
    }
    
    private void openDoors() {
        System.out.println("Elevator " + id + " doors opening at floor " + currentFloor);
    }
    
    private void closeDoors() {
        System.out.println("Elevator " + id + " doors closing at floor " + currentFloor);
    }
    
    public int distanceToFloor(int floor) {
        return Math.abs(currentFloor - floor);
    }
}

class ElevatorScheduler {
    private List<Elevator> elevators;
    
    public ElevatorScheduler(List<Elevator> elevators) {
        this.elevators = elevators;
    }
    
    public void scheduleRequest(ExternalRequest request) {
        Elevator bestElevator = findBestElevator(request);
        InternalRequest internalReq = new InternalRequest(
            request.getFloor(), 
            request.getDirection()
        );
        bestElevator.addRequest(internalReq);
    }
    
    private Elevator findBestElevator(ExternalRequest request) {
        // SCAN algorithm: prefer elevators moving in same direction
        Elevator best = null;
        int minDistance = Integer.MAX_VALUE;
        
        for (Elevator elevator : elevators) {
            if (elevator.getStatus() == ElevatorStatus.MAINTENANCE) {
                continue;
            }
            
            // Same direction or idle
            if (elevator.getDirection() == request.getDirection() || 
                elevator.getDirection() == Direction.IDLE) {
                
                int distance = elevator.distanceToFloor(request.getFloor());
                if (distance < minDistance) {
                    minDistance = distance;
                    best = elevator;
                }
            }
        }
        
        // If no elevator in same direction, pick closest idle
        if (best == null) {
            best = elevators.stream()
                .filter(e -> e.getDirection() == Direction.IDLE)
                .min(Comparator.comparingInt(
                    e -> e.distanceToFloor(request.getFloor())
                ))
                .orElse(elevators.get(0));
        }
        
        return best;
    }
}

class ExternalRequest {
    private int floor;
    private Direction direction;
    
    public ExternalRequest(int floor, Direction direction) {
        this.floor = floor;
        this.direction = direction;
    }
}

class InternalRequest {
    private int floor;
    private Direction direction;
    
    public InternalRequest(int floor, Direction direction) {
        this.floor = floor;
        this.direction = direction;
    }
}

class RequestComparator implements Comparator<InternalRequest> {
    @Override
    public int compare(InternalRequest r1, InternalRequest r2) {
        return Integer.compare(r1.getFloor(), r2.getFloor());
    }
}
```

### Scheduling Algorithms
1. **FCFS**: First Come First Serve (simple but inefficient)
2. **SCAN**: Move in one direction, serve all requests, then reverse
3. **LOOK**: Like SCAN but reverse when no more requests ahead
4. **Destination Dispatch**: Assign elevator before boarding

### Key Considerations
- Load balancing across elevators
- Energy efficiency (minimize movement)
- Peak hour handling
- Emergency protocols

